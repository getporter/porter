package claims

import (
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/schema"
)

var _ storage.Document = Run{}

// Run represents the execution of an installation's bundle.
type Run struct {
	// SchemaVersion of the document.
	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`

	// ID of the Run.
	ID string `json:"_id" yaml:"_id", toml:"_id"`

	// Created timestamp of the Run.
	Created time.Time `json:"created" yaml:"created", toml:"created"`

	// Namespace of the installation.
	Namespace string `json:"namespace" yaml:"namespace", toml:"namespace"`

	// Installation name.
	Installation string `json:"installation" yaml:"installation", toml:"installation"`

	// Revision of the installation.
	Revision string `json:"revision" yaml:"revision", toml:"revision"`

	// Action executed against the installation.
	Action string `json:"action" yaml:"action", toml:"action"`

	// Bundle is the definition of the bundle.
	Bundle bundle.Bundle `json:"bundle" yaml:"bundle", toml:"bundle"`

	// BundleReference is the canonical reference to the bundle used in the action.
	BundleReference string `json:"bundleReference" yaml:"bundleReference", toml:"bundleReference"`

	// BundleDigest is the digest of the bundle.
	// TODO(carolynvs): populate this
	BundleDigest string `json:"bundleDigest" yaml:"bundleDigest", toml:"bundleDigest"`

	// Parameters are the key/value pairs that were passed in during the operation.
	Parameters map[string]interface{} `json:"parameters" yaml:"parameters", toml:"parameters"`

	// Custom extension data applicable to a given runtime.
	// TODO(carolynvs): remove custom and populate it in ToCNAB
	Custom interface{} `json:"custom" yaml:"custom", toml:"custom"`
}

func (r Run) DefaultDocumentFilter() interface{} {
	return map[string]interface{}{"_id": r.ID}
}

// NewRun creates a run with default values initialized.
func NewRun(namespace string, installation string) Run {
	return Run{
		SchemaVersion: SchemaVersion,
		ID:            cnab.NewULID(),
		Revision:      cnab.NewULID(),
		Created:       time.Now(),
		Namespace:     namespace,
		Installation:  installation,
	}
}

// ShouldRecord the current run in the Installation history.
// Runs are only recorded for actions that modify the bundle resources,
// or for stateful actions. Stateless actions do not require an existing
// installation or credentials, and are for actions such as documentation, dry-run, etc.
func (r Run) ShouldRecord() bool {
	stateless := false
	if customAction, ok := r.Bundle.Actions[r.Action]; ok {
		stateless = customAction.Stateless
	}
	modifies, _ := r.IsModifyingAction()
	return modifies || !stateless
}

// IsModifyingAction returns if the specified action modifies bundle resources.
func (r Run) IsModifyingAction() (bool, error) {
	action, err := r.Bundle.GetAction(r.Action)
	return action.Modifies, err
}

// ToCNAB associated with the Run.
func (r Run) ToCNAB() cnab.Claim {
	return cnab.Claim{
		SchemaVersion:   CNABSchemaVersion(),
		ID:              r.ID,
		Namespace:       r.Namespace,
		Installation:    r.Installation,
		Revision:        r.Revision,
		Created:         r.Created,
		Action:          r.Action,
		Bundle:          r.Bundle,
		BundleReference: r.BundleReference,
		BundleDigest:    r.BundleDigest,
		Parameters:      r.Parameters,
		Custom:          r.Custom,
	}
}

// NewRun creates a result for the current Run.
func (r Run) NewResult(status string) Result {
	result := NewResult()
	result.RunID = r.ID
	result.Namespace = r.Namespace
	result.Installation = r.Installation
	result.Status = status
	return result
}

// NewResultFrom creates a result from the output of a CNAB run.
func (r Run) NewResultFrom(cnabResult cnab.Result) Result {
	return Result{
		SchemaVersion:  SchemaVersion,
		ID:             cnabResult.ID,
		Namespace:      r.Namespace,
		Installation:   r.Installation,
		RunID:          r.ID,
		Created:        cnabResult.Created,
		Status:         cnabResult.Status,
		Message:        cnabResult.Message,
		OutputMetadata: cnabResult.OutputMetadata,
		Custom:         cnabResult.Custom,
	}
}
