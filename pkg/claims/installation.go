package claims

import (
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/Masterminds/semver/v3"
	"github.com/cnabio/cnab-go/schema"
	"github.com/opencontainers/go-digest"
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

	// BundleTag is the OCI tag of the current bundle definition.
	BundleTag string `json:"bundleTag,omitempty" yaml:"bundleTag,omitempty" toml:"bundleTag,omitempty"`

	// Custom extension data applicable to a given runtime.
	// TODO(carolynvs): remove and populate in ToCNAB when we firm up the spec
	Custom interface{} `json:"custom,omitempty" yaml:"custom,omitempty" toml:"custom,omitempty"`

	// Labels applied to the installation.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" toml:"labels,omitempty"`

	// Parameters specified by the user through overrides.
	// Does not include defaults, or values resolved from parameter sources.
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty" toml:"parameters,omitempty"`

	// CredentialSets that should be included when the bundle is reconciled.
	CredentialSets []string `json:"credentialSets,omitempty" yaml:"credentialSets,omitempty" toml:"credentialSets,omitempty"`

	// ParameterSets that should be included when the bundle is reconciled.
	ParameterSets []string `json:"parameterSets,omitempty" yaml:"parameterSets,omitempty" toml:"parameterSets,omitempty"`

	// Status of the installation.
	Status InstallationStatus `json:"status,omitempty" yaml:"status,omitempty" toml:"status,omitempty"`
}

func (i Installation) String() string {
	return fmt.Sprintf("%s/%s", i.Namespace, i.Name)
}

func (i Installation) GetBundleReference() (cnab.OCIReference, bool, error) {
	if i.BundleRepository == "" {
		return cnab.OCIReference{}, false, nil
	}

	ref, err := cnab.ParseOCIReference(i.BundleRepository)
	if err != nil {
		return cnab.OCIReference{}, false, errors.Wrapf(err, "invalid BundleRepository %s", i.BundleRepository)
	}

	if i.BundleDigest != "" {
		d, err := digest.Parse(i.BundleDigest)
		if err != nil {
			return cnab.OCIReference{}, false, errors.Wrapf(err, "invalid BundleDigest %s", i.BundleDigest)
		}

		ref, err = ref.WithDigest(d)
		if err != nil {
			return cnab.OCIReference{}, false, errors.Wrapf(err, "error joining the BundleRepository %s and BundleDigest %s", i.BundleRepository, i.BundleDigest)
		}
		return ref, true, nil
	}

	if i.BundleVersion != "" {
		v, err := semver.NewVersion(i.BundleVersion)
		if err != nil {
			return cnab.OCIReference{}, false, errors.New("invalid BundleVersion")
		}

		// The bundle version feature can only be used with standard naming conventions
		// everyone else can use the tag field if they do weird things
		ref, err = ref.WithTag("v" + v.String())
		if err != nil {
			return cnab.OCIReference{}, false, errors.Wrapf(err, "error joining the BundleRepository %s and BundleVersion %s", i.BundleRepository, i.BundleVersion)
		}
		return ref, true, nil
	}

	if i.BundleTag != "" {
		ref, err = ref.WithTag(i.BundleTag)
		if err != nil {
			return cnab.OCIReference{}, false, errors.Wrapf(err, "error joining the BundleRepository %s and BundleTag %s", i.BundleRepository, i.BundleTag)
		}
		return ref, true, nil
	}

	return cnab.OCIReference{}, false, errors.New("Invalid installation, either BundleDigest, BundleVersion or BundleTag must be specified")
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
		i.Status.BundleReference = run.BundleReference
		i.Status.BundleVersion = run.Bundle.Version
		i.Status.BundleDigest = run.BundleDigest
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
	i.BundleTag = input.BundleTag
	i.Parameters = input.Parameters
	i.CredentialSets = input.CredentialSets
	i.ParameterSets = input.ParameterSets
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
	_, _, err := i.GetBundleReference()
	return errors.Wrapf(err, "could not determine the fully-qualified bundle reference")
}

// TrackBundle updates the bundle that the installation is tracking.
func (i *Installation) TrackBundle(ref cnab.OCIReference) {
	// Determine if the bundle is managed by version, digest or tag
	i.BundleRepository = ref.Repository()
	if ref.HasVersion() {
		i.BundleVersion = ref.Version()
	} else if ref.HasDigest() {
		i.BundleDigest = ref.Digest().String()
	} else {
		i.BundleTag = ref.Tag()
	}
}

// SetLabel on the installation.
func (i *Installation) SetLabel(key string, value string) {
	if i.Labels == nil {
		i.Labels = make(map[string]string, 1)
	}
	i.Labels[key] = value
}

// ConvertParameterValues converts each parameter from an unknown type to
// the type specified for that parameter on the bundle.
func (i *Installation) ConvertParameterValues(b cnab.ExtendedBundle) error {
	for paramName, rawParamValue := range i.Parameters {
		typedValue, err := b.ConvertParameterValue(paramName, rawParamValue)
		if err != nil {
			return err
		}
		i.Parameters[paramName] = typedValue
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

	// BundleReference of the bundle that last altered the installation state.
	BundleReference string `json:"bundleReference" yaml:"bundleReference" toml:"bundleReference"`

	// BundleVersion is the version of the bundle that last altered the installation state.
	BundleVersion string `json:"bundleVersion" yaml:"bundleVersion" toml:"bundleVersion"`

	// BundleDigest is the digest of the bundle that last altered the installation state.
	BundleDigest string `json:"bundleDigest" yaml:"bundleDigest" toml:"bundleDigest"`
}
