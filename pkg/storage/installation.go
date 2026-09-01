package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/Masterminds/semver/v3"
	"github.com/opencontainers/go-digest"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ Document = Installation{}

type Installation struct {
	// ID is the unique identifier for an installation record.
	ID string `json:"id"`

	InstallationSpec

	// Status of the installation.
	Status InstallationStatus `json:"status,omitempty"`
}

// InstallationSpec contains installation fields that represent the desired state of the installation.
type InstallationSpec struct {
	// SchemaType indicates the type of resource imported from a file.
	SchemaType string `json:"schemaType"`

	// SchemaVersion is the version of the installation state schema.
	SchemaVersion cnab.SchemaVersion `json:"schemaVersion"`

	// Name of the installation. Immutable.
	Name string `json:"name"`

	// Namespace in which the installation is defined.
	Namespace string `json:"namespace"`

	// Uninstalled specifies if the installation isn't used anymore and should be uninstalled.
	Uninstalled bool `json:"uninstalled,omitempty"`

	// Bundle specifies the bundle reference to use with the installation.
	Bundle OCIReferenceParts `json:"bundle"`

	// Custom extension data applicable to a given runtime.
	// TODO(carolynvs): remove and populate in ToCNAB when we firm up the spec
	Custom interface{} `json:"custom,omitempty"`

	// Labels applied to the installation.
	Labels map[string]string `json:"labels,omitempty"`

	// CredentialSets that should be included when the bundle is reconciled.
	CredentialSets []string `json:"credentialSets,omitempty"`

	// ParameterSets that should be included when the bundle is reconciled.
	ParameterSets []string `json:"parameterSets,omitempty"`

	// Parameters specified by the user through overrides.
	// Does not include defaults, or values resolved from parameter sources.
	Parameters ParameterSet `json:"parameters,omitempty"`
}

func (i InstallationSpec) String() string {
	return fmt.Sprintf("%s/%s", i.Namespace, i.Name)
}

func (i Installation) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"namespace": i.Namespace, "name": i.Name}
}

func NewInstallation(namespace string, name string) Installation {
	now := time.Now()
	return Installation{
		ID: cnab.NewULID(),
		InstallationSpec: InstallationSpec{
			SchemaType:    SchemaTypeInstallation,
			SchemaVersion: DefaultInstallationSchemaVersion,
			Namespace:     namespace,
			Name:          name,
			Parameters:    NewInternalParameterSet(namespace, name),
		},
		Status: InstallationStatus{
			Created:  now,
			Modified: now,
		},
	}
}

// NewRun creates a run of the current bundle.
func (i Installation) NewRun(action string, b cnab.ExtendedBundle) Run {
	run := NewRun(i.Namespace, i.Name)
	run.Action = action

	// Copy over relevant overrides from the installation to the run
	// An installation may have an overridden parameter that doesn't apply to this current action
	run.ParameterOverrides = NewInternalParameterSet(i.Namespace, i.Name)
	for _, p := range i.Parameters.Parameters {
		if parmDef, ok := b.Parameters[p.Name]; ok {
			if !parmDef.AppliesTo(action) {
				continue
			}
			run.ParameterOverrides.Parameters = append(run.ParameterOverrides.Parameters, p)
		}
	}

	return run
}

// ApplyResult updates cached status data on the installation from the
// last bundle run.
func (i *Installation) ApplyResult(run Run, result Result) {
	// Update the installation with the last modifying action
	if action, err := run.Bundle.GetAction(run.Action); err == nil && action.Modifies {
		i.Status.BundleReference = run.BundleReference
		i.Status.BundleVersion = run.Bundle.Version
		i.Status.BundleDigest = run.BundleDigest
		i.Status.RunID = run.ID
		i.Status.Action = run.Action
		i.Status.ResultID = result.ID
		i.Status.ResultStatus = result.Status
	}

	if !i.IsInstalled() && run.Action == cnab.ActionInstall && result.Status == cnab.StatusSucceeded {
		i.Status.Installed = &result.Created
	}

	if !i.IsUninstalled() && run.Action == cnab.ActionUninstall && result.Status == cnab.StatusSucceeded {
		i.Status.Uninstalled = &result.Created
	}
}

// Apply user-provided changes to an existing installation.
// Only updates fields that users are allowed to modify.
// For example, Name, Namespace and Status cannot be modified.
func (i *InstallationSpec) Apply(input InstallationSpec) {
	i.SchemaType = input.SchemaType
	i.SchemaVersion = input.SchemaVersion
	i.Uninstalled = input.Uninstalled
	i.Bundle = input.Bundle
	i.Parameters = input.Parameters
	i.CredentialSets = input.CredentialSets
	i.ParameterSets = input.ParameterSets
	i.Labels = input.Labels
}

// Validate the installation document and report the first error.
func (i *InstallationSpec) Validate(ctx context.Context, strategy schema.CheckStrategy) error {
	_, span := tracing.StartSpan(ctx,
		attribute.String("installation", i.String()),
		attribute.String("schemaVersion", string(i.SchemaVersion)),
		attribute.String("defaultSchemaVersion", string(DefaultInstallationSchemaVersion)))
	defer span.EndSpan()

	// Before we can validate, get our resource in a consistent state
	// 1. Check if we know what to do with this version of the resource
	defaultSchemaVersion := semver.MustParse(string(DefaultInstallationSchemaVersion))
	if warnOnly, err := schema.ValidateSchemaVersion(strategy, SupportedInstallationSchemaVersions, string(i.SchemaVersion), defaultSchemaVersion); err != nil {
		if warnOnly {
			span.Warn(err.Error())
		} else {
			return span.Error(err)
		}
	}

	// 2. Check if they passed in the right resource type
	if i.SchemaType != "" && !strings.EqualFold(i.SchemaType, SchemaTypeInstallation) {
		return span.Errorf("invalid schemaType %s, expected %s", i.SchemaType, SchemaTypeInstallation)
	}

	// OK! Now we can do resource specific validations

	// Default the schema type before importing into the database if it's not set already
	// SchemaType isn't really used by our code, it's a type hint for editors, but this will ensure we are consistent in our persisted documents
	if i.SchemaType == "" {
		i.SchemaType = SchemaTypeInstallation
	}

	// OK! Now we can do resource specific validations

	// We can change these to better checks if we consolidate our logic around the various ways we let you
	// install from a bundle definition https://github.com/getporter/porter/issues/1024#issuecomment-899828081
	// Until then, these are pretty weak checks
	_, _, err := i.Bundle.GetBundleReference()
	if err != nil {
		return span.Errorf("could not determine the fully-qualified bundle reference: %w", err)
	}

	return nil
}

// TrackBundle updates the bundle that the installation is tracking.
func (i *Installation) TrackBundle(ref cnab.OCIReference) {
	// Determine if the bundle is managed by version, digest or tag
	i.Bundle.Repository = ref.Repository()
	if ref.HasVersion() {
		i.Bundle.Version = ref.Version()
	} else if ref.HasDigest() {
		i.Bundle.Digest = ref.Digest().String()
	} else {
		i.Bundle.Tag = ref.Tag()
	}
}

// SetLabel on the installation.
func (i *Installation) SetLabel(key string, value string) {
	if i.Labels == nil {
		i.Labels = make(map[string]string, 1)
	}
	i.Labels[key] = value
}

// NewInternalParameterSet creates a new ParameterSet that's used to store
// parameter overrides with the required fields initialized.
func (i Installation) NewInternalParameterSet(params ...secrets.SourceMap) ParameterSet {
	return NewInternalParameterSet(i.Namespace, i.ID, params...)
}

func (i Installation) AddToTrace(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	doc, _ := json.Marshal(i)
	span.SetAttributes(
		attribute.String("installation", i.String()),
		attribute.String("installationDefinition", string(doc)))
}

// InstallationStatus's purpose is to assist with making porter list be able to display everything
// with a single database query. Do not replicate data available on Run and Result here.
type InstallationStatus struct {
	// RunID of the bundle execution that last altered the installation status.
	RunID string `json:"runId" yaml:"runId" toml:"runId"`

	// Action of the bundle run that last informed the installation status.
	Action string `json:"action" yaml:"action" toml:"action"`

	// ResultID of the result that last informed the installation status.
	ResultID string `json:"resultId" yaml:"resultId" toml:"resultId"`

	// ResultStatus is the status of the result that last informed the installation status.
	ResultStatus string `json:"resultStatus" yaml:"resultStatus" toml:"resultStatus"`

	// Created timestamp of the installation.
	Created time.Time `json:"created" yaml:"created" toml:"created"`

	// Modified timestamp of the installation.
	Modified time.Time `json:"modified" yaml:"modified" toml:"modified"`

	// Installed indicates if the install action has successfully completed for this installation.
	// Once that state is reached, Porter should not allow it to be reinstalled as a protection from installations
	// being overwritten.
	Installed *time.Time `json:"installed" yaml:"installed" toml:"installed"`

	// Uninstalled indicates if the installation has successfully completed the uninstall action.
	// Once that state is reached, Porter should not allow further stateful actions.
	Uninstalled *time.Time `json:"uninstalled" yaml:"uninstalled" toml:"uninstalled"`

	// BundleReference of the bundle that last altered the installation state.
	BundleReference string `json:"bundleReference" yaml:"bundleReference" toml:"bundleReference"`

	// BundleVersion is the version of the bundle that last altered the installation state.
	BundleVersion string `json:"bundleVersion" yaml:"bundleVersion" toml:"bundleVersion"`

	// BundleDigest is the digest of the bundle that last altered the installation state.
	BundleDigest string `json:"bundleDigest" yaml:"bundleDigest" toml:"bundleDigest"`

	// OutputPersistFailed indicates that one or more outputs from the run that
	// last informed this status failed to persist to the secret store. The
	// installation itself may still have succeeded.
	OutputPersistFailed bool `json:"outputPersistFailed,omitempty" yaml:"outputPersistFailed,omitempty" toml:"outputPersistFailed,omitempty"`

	// References tracks other installations that depend on this installation
	// to satisfy one of their dependencies, either because it was created for
	// or reused to satisfy that dependency. Used to determine if it's safe to
	// delete this installation.
	References []InstallationReference `json:"references,omitempty" yaml:"references,omitempty" toml:"references,omitempty"`
}

// InstallationReference records another installation that depends on this
// installation to satisfy one of its dependencies.
type InstallationReference struct {
	// Installation is the "namespace/name" of the referencing installation.
	Installation string `json:"installation" yaml:"installation" toml:"installation"`

	// Dependency is the alias the referencing installation's bundle uses for this dependency.
	Dependency string `json:"dependency" yaml:"dependency" toml:"dependency"`
}

// AddReference records that installation depends on i to satisfy its
// dependency. Idempotent: upserts by (installation, dependency), returning
// true only if it actually changed References, so callers can skip a
// needless persist. Callers are responsible for read-modify-write
// consistency; there's no built-in optimistic concurrency here, matching
// the rest of this file.
func (i *Installation) AddReference(installation string, dependency string) bool {
	for _, ref := range i.Status.References {
		if ref.Installation == installation && ref.Dependency == dependency {
			return false
		}
	}

	i.Status.References = append(i.Status.References, InstallationReference{
		Installation: installation,
		Dependency:   dependency,
	})
	return true
}

// RemoveReference removes every reference recorded for the given
// referencing installation (all of its dependency aliases at once), e.g.
// when that installation is deleted or no longer needs this installation.
// Returns true only if it actually changed References.
func (i *Installation) RemoveReference(installation string) bool {
	kept := i.Status.References[:0]
	changed := false
	for _, ref := range i.Status.References {
		if ref.Installation == installation {
			changed = true
			continue
		}
		kept = append(kept, ref)
	}
	i.Status.References = kept
	return changed
}

// IsReferenced reports whether any other installation currently depends on
// this installation.
func (i Installation) IsReferenced() bool {
	return len(i.Status.References) > 0
}

// IsInstalled checks if the installation is currently installed.
func (i Installation) IsInstalled() bool {
	if i.Status.Uninstalled != nil && i.Status.Installed != nil {
		return i.Status.Installed.After(*i.Status.Uninstalled)
	}
	return i.Status.Uninstalled == nil && i.Status.Installed != nil
}

// IsUninstalled checks if the installation has been uninstalled.
func (i Installation) IsUninstalled() bool {
	if i.Status.Uninstalled != nil && i.Status.Installed != nil {
		return i.Status.Uninstalled.After(*i.Status.Installed)
	}
	return i.Status.Uninstalled != nil
}

// IsDefined checks if the installation is has already been defined but not installed yet.
func (i Installation) IsDefined() bool {
	return i.Status.Installed == nil
}

// OCIReferenceParts is our storage representation of cnab.OCIReference
// with the parts explicitly stored separately so that they are queryable.
type OCIReferenceParts struct {
	// Repository is the OCI repository of the bundle.
	// For example, "getporter/porter-hello".
	Repository string `json:"repository,omitempty" yaml:"repository,omitempty" toml:"repository,omitempty"`

	// Version is the current version of the bundle.
	// For example, "1.2.3".
	Version string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`

	// Digest is the current digest of the bundle.
	// For example, "sha256:abc123"
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty" toml:"digest,omitempty"`

	// Tag is the OCI tag of the bundle.
	// For example, "latest".
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty" toml:"tag,omitempty"`
}

func (r OCIReferenceParts) GetBundleReference() (cnab.OCIReference, bool, error) {
	if r.Repository == "" {
		return cnab.OCIReference{}, false, nil
	}

	ref, err := cnab.ParseOCIReference(r.Repository)
	if err != nil {
		return cnab.OCIReference{}, false, fmt.Errorf("invalid bundle Repository %s: %w", r.Repository, err)
	}

	if r.Digest != "" {
		d, err := digest.Parse(r.Digest)
		if err != nil {
			return cnab.OCIReference{}, false, fmt.Errorf("invalid bundle Digest %s: %w", r.Digest, err)
		}

		ref, err = ref.WithDigest(d)
		if err != nil {
			return cnab.OCIReference{}, false, fmt.Errorf("error joining the bundle Repository %s and Digest %s: %w", r.Repository, r.Digest, err)
		}
		return ref, true, nil
	}

	if r.Version != "" {
		// The bundle version feature can only be used with standard naming conventions
		// everyone else can use the tag field if they do weird things
		ref, err = ref.WithVersion(r.Version)
		if err != nil {
			return cnab.OCIReference{}, false, fmt.Errorf("error joining the bundle Repository %s and Version %s: %w", r.Repository, r.Version, err)
		}
		return ref, true, nil
	}

	if r.Tag != "" {
		ref, err = ref.WithTag(r.Tag)
		if err != nil {
			return cnab.OCIReference{}, false, fmt.Errorf("error joining the bundle Repository %s and Tag %s: %w", r.Repository, r.Tag, err)
		}
		return ref, true, nil
	}

	return cnab.OCIReference{}, false, errors.New("Invalid bundle reference, either Digest, Version, or Tag must be specified")
}
