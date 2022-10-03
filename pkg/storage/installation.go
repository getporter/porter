package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/Masterminds/semver/v3"
	"github.com/cnabio/cnab-go/schema"
	"github.com/opencontainers/go-digest"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ Document = Installation{}

type Installation struct {
	// SchemaVersion is the version of the installation state schema.
	SchemaVersion schema.Version `json:"schemaVersion"`

	// ID is the unique identifier for an installation record.
	ID string `json:"id"`

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

	// Status of the installation.
	Status InstallationStatus `json:"status,omitempty"`
}

func (i Installation) String() string {
	return fmt.Sprintf("%s/%s", i.Namespace, i.Name)
}

func (i Installation) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"namespace": i.Namespace, "name": i.Name}
}

func NewInstallation(namespace string, name string) Installation {
	now := time.Now()
	return Installation{
		SchemaVersion: InstallationSchemaVersion,
		ID:            cnab.NewULID(),
		Namespace:     namespace,
		Name:          name,
		Parameters:    NewInternalParameterSet(namespace, name),
		Status: InstallationStatus{
			Created:  now,
			Modified: now,
		},
	}
}

// NewRun creates a run of the current bundle.
func (i Installation) NewRun(action string) Run {
	run := NewRun(i.Namespace, i.Name)
	run.Action = action
	run.ParameterOverrides = i.Parameters
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
func (i *Installation) Apply(input Installation) {
	i.Uninstalled = input.Uninstalled
	i.Bundle = input.Bundle
	i.Parameters = input.Parameters
	i.CredentialSets = input.CredentialSets
	i.ParameterSets = input.ParameterSets
	i.Labels = input.Labels
}

// Validate the installation document and report the first error.
func (i *Installation) Validate() error {
	if InstallationSchemaVersion != i.SchemaVersion {
		if i.SchemaVersion == "" {
			i.SchemaVersion = "(none)"
		}
		return fmt.Errorf("invalid schemaVersion provided: %s. This version of Porter is compatible with %s.", i.SchemaVersion, InstallationSchemaVersion)
	}

	// We can change these to better checks if we consolidate our logic around the various ways we let you
	// install from a bundle definition https://github.com/getporter/porter/issues/1024#issuecomment-899828081
	// Until then, these are pretty weak checks
	_, _, err := i.Bundle.GetBundleReference()
	if err != nil {
		return fmt.Errorf("could not determine the fully-qualified bundle reference: %w", err)
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
func (i Installation) NewInternalParameterSet(params ...secrets.Strategy) ParameterSet {
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
		v, err := semver.NewVersion(r.Version)
		if err != nil {
			return cnab.OCIReference{}, false, errors.New("invalid BundleVersion")
		}

		// The bundle version feature can only be used with standard naming conventions
		// everyone else can use the tag field if they do weird things
		ref, err = ref.WithTag("v" + v.String())
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
