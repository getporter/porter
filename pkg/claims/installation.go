package claims

import (
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/Masterminds/semver/v3"
	"github.com/cnabio/cnab-go/schema"
	"github.com/pkg/errors"
)

const (
	// SchemaVersion represents the version associated with the schema
	// for all installation documents: installations, runs, results and outputs.
	SchemaVersion = schema.Version("1.0.0")
)

var _ storage.Document = Installation{}

type Installation struct {
	// SchemaVersion is the version of the installation state schema.
	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`

	// Name of the installation. Immutable.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Namespace in which the installation is defined.
	Namespace string `json:"namespace" yaml:"namespace" toml:"namespace"`

	// Created timestamp of the installation.
	Created time.Time `json:"created" yaml:"created" toml:"created"`

	// Modified timestamp of the installation.
	Modified time.Time `json:"modified" yaml:"modified" toml:"modified"`

	// BundleRepository is the OCI repository of the current bundle definition.
	BundleRepository string `json:"bundleRepository,omitempty" yaml:"bundleRepository,omitempty" toml:"bundleRepository,omitempty"`

	// BundleVersion is the current version of the bundle.
	BundleVersion string `json:"bundleVersion,omitempty" yaml:"bundleVersion,omitempty" toml:"bundleVersion,omitempty"`

	// BundleDigest is the current digest of the bundle.
	BundleDigest string `json:"bundleDigest,omitempty" yaml:"bundleDigest,omitempty" toml:"bundleDigest,omitempty"`

	// Custom extension data applicable to a given runtime.
	// TODO(carolynvs): remove and populate in ToCNAB when we firm up the spec
	Custom interface{} `json:"custom,omitempty" yaml:"custom,omitempty" toml:"custom,omitempty"`

	// Labels applied to the installation.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" toml:"labels,omitempty"`

	// Parameters specified by the user through overrides or parameter sets.
	// Does not include defaults, or values resolved from parameter sources.
	Parameters map[string]interface{} `json:"parameters" yaml:"parameters" toml:"parameters"`

	// CredentialSets that should be included when the bundle is reconciled.
	CredentialSets []string `json:"credentialSets" yaml:"credentialSets" toml:"credentialSets"`

	// ParameterSets that should be included when the bundle is reconciled.
	ParameterSets []string `json:"parameterSets" yaml:"parameterSets" toml:"parameterSets"`

	// Status of the installation.
	Status InstallationStatus `json:"status" yaml:"status" toml:"status"`
}

func (i Installation) String() string {
	return fmt.Sprintf("%s/%s", i.Namespace, i.Name)
}

func (i Installation) DefaultDocumentFilter() interface{} {
	return map[string]interface{}{"namespace": i.Namespace, "name": i.Name}
}

func NewInstallation(namespace string, name string) Installation {
	now := time.Now()
	return Installation{
		SchemaVersion: SchemaVersion,
		Namespace:     namespace,
		Name:          name,
		Created:       now,
		Modified:      now,
	}
}

func (i Installation) ToCNAB() cnab.Installation {
	// TODO(carolynvs): Remove installation status from the cnab struct
	// in general look over what is actually needed to be specified on an installation doc. Does it need to be in the spec?
	return cnab.Installation{
		SchemaVersion:    CNABSchemaVersion(),
		Name:             i.Name,
		Namespace:        i.Namespace,
		BundleRepository: i.BundleRepository,
		BundleVersion:    i.BundleVersion,
		BundleDigest:     i.BundleDigest,
		Created:          i.Created,
		Modified:         i.Modified,
		Custom:           i.Custom,
		Labels:           i.Labels,
	}
}

// NewRun creates a run of the current bundle.
func (i Installation) NewRun(action string) Run {
	run := NewRun(i.Namespace, i.Name)
	run.Action = action
	return run
}

// ApplyResult updates cached status data on the installation from the
// last bundle run.
func (i *Installation) ApplyResult(run Run, result Result) {
	// Update the installation with the last modifying action
	if action, err := run.Bundle.GetAction(run.Action); err == nil && action.Modifies {
		i.Status.RunID = run.ID
		i.Status.Action = run.Action
		i.Status.ResultID = result.ID
		i.Status.ResultStatus = result.Status
	}

	if !i.Status.InstallationCompleted && run.Action == cnab.ActionInstall && result.Status == cnab.StatusSucceeded {
		i.Status.InstallationCompleted = true
	}
}

// Apply user-provided changes to an existing installation.
// Only updates fields that users are allowed to modify.
// For example, Name, Namespace and Status cannot be modified.
func (i *Installation) Apply(input Installation) {
	i.BundleRepository = input.BundleRepository
	i.BundleDigest = input.BundleDigest
	i.BundleVersion = input.BundleVersion
	i.Parameters = input.Parameters
	i.Labels = input.Labels
}

// Validate the installation document and report the first error.
func (i *Installation) Validate() error {
	if SchemaVersion != i.SchemaVersion {
		return errors.Errorf("invalid schemaVersion provided: %s. This version of Porter is compatible with %s.", i.SchemaVersion, SchemaVersion)
	}

	// We can change these to better checks if we consolidate our logic around the various ways we let you
	// install from a bundle definition https://github.com/getporter/porter/issues/1024#issuecomment-899828081
	// Until then, these are pretty weak checks

	if i.BundleRepository != "" {
		ref, err := cnab.ParseOCIReference(i.BundleRepository)
		if err != nil {
			return errors.Wrapf(err, "invalid bundleRepository %s", i.BundleRepository)
		}
		i.BundleRepository = ref.Repository()
	}

	if i.BundleVersion != "" {
		if _, err := semver.NewVersion(i.BundleVersion); err != nil {
			return errors.Wrapf(err, "Invalid bundleVersion. Must be a valid v2 semver value.")
		}
	}

	if i.BundleRepository != "" && i.BundleDigest == "" && i.BundleVersion == "" {
		return errors.Errorf("either the bundleVersion or the bundleDigest must be provided.")
	}

	return nil
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

	// InstallationCompleted indicates if the install action has successfully completed for this installation.
	// Once that state is reached, Porter should not allow it to be reinstalled as a protection from installations
	// being overwritten.
	InstallationCompleted bool `json:"installationCompleted" yaml:"installationCompleted" toml:"installationCompleted"`
}
