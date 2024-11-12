package storage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cnabio/cnab-go/bundle"

	"get.porter.sh/porter/pkg/cnab"
)

var _ Document = Run{}
var _ json.Marshaler = Run{}
var _ json.Unmarshaler = &Run{}

// Run represents the execution of an installation's bundle. It contains both the
// instructions used by Porter to run the bundle, and additional status/audit
// fields so users can keep track of how the bundle was run.
type Run struct {
	// SchemaVersion of the document.
	SchemaVersion cnab.SchemaVersion `json:"schemaVersion"`

	// ID of the Run.
	ID string `json:"_id"`

	// Created timestamp of the Run.
	Created time.Time `json:"created"`

	// Modified timestamp of the Run, set when we resolve run parameters just-in-time.
	// A run can be created ahead of time as Pending and not have its parameters resolved until much later.
	Modified time.Time `json:"modified"`

	// Namespace of the installation.
	Namespace string `json:"namespace"`

	// Installation name.
	Installation string `json:"installation"`

	// Revision of the installation.
	Revision string `json:"revision"`

	// Action executed against the installation.
	Action string `json:"action"`

	// Bundle is the definition of the bundle.
	// Bundle has custom marshal logic in MarshalJson.
	Bundle bundle.Bundle `json:"-" gorm:"json"`

	// BundleReference is the canonical reference to the bundle used in the action.
	BundleReference string `json:"bundleReference"`

	// BundleDigest is the digest of the bundle.
	// TODO(carolynvs): populate this
	BundleDigest string `json:"bundleDigest"`

	// ParameterOverrides are the key/value parameter overrides (taking precedence over
	// parameters specified in a parameter set) specified during the run.
	// This is a status/audit field and is not used to resolve parameters for a Run.
	ParameterOverrides ParameterSet `json:"parameterOverrides,omitempty" gorm:"embedded;embeddedPrefix:parameterOverrides_"`

	// CredentialSets is a list of the credential set names used during the run.
	// This is a status/audit field and is not used to resolve credentials for a Run.
	CredentialSets []string `json:"credentialSets,omitempty" gorm:"type:jsonb"`

	// ParameterSets is the list of parameter set names used during the run.
	// This is a status/audit field and is not used to resolve parameters for a Run.
	ParameterSets []string `json:"parameterSets,omitempty" gorm:"type:jsonb"`

	// Parameters is the full set of parameters that should be resolved just-in-time
	// (JIT) before executing the bundle. This includes internal parameters,
	// parameter sources, values from parameter sets, etc. These should be a "clean"
	// set of parameters that have sensitive values persisted in secrets using the
	// Sanitizer.
	// After the parameters are resolved, this structure holds (but does not marshal)
	// the resolved values, in addition to the mapping strategy.
	Parameters ParameterSet `json:"parameters,omitempty" gorm:"embedded;embeddedPrefix:parameters_"`

	// Custom extension data applicable to a given runtime.
	// TODO(carolynvs): remove custom and populate it in ToCNAB
	Custom interface{} `json:"custom" gorm:"type:jsonb"`

	// ParametersDigest is a hash or digest of the final set of parameters, which allows us to
	// quickly determine if the parameters have changed without requiring that they
	// are re-resolved. The value should contain the hash type, e.g. sha256:abc123...
	// This is a status/audit field and is not used to resolve parameters for a Run.
	ParametersDigest string `json:"parametersDigest,omitempty"`

	// Credentials is the full set of credentials that should be resolved
	// just-in-time (JIT) before executing the bundle. These should be a "clean" set
	// of parameters that have sensitive values persisted in secrets using the
	// Sanitizer.
	Credentials CredentialSet `json:"credentials,omitempty" gorm:"embedded;embeddedPrefix:credentials_"`

	// CredentialsDigest is a hash or digest of the final set of credentials, which allows us to
	// quickly determine if the credentials have changed without requiring that they
	// are re-resolved. The value should contain the hash type, e.g. sha256:abc123...
	// This is a status/audit field and is not used to resolve credentials for a Run.
	CredentialsDigest string `json:"credentialsDigest,omitempty"`
}

// rawRun is an alias for Run that does not have a json marshal functions defined,
// so it's safe to marshal without causing infinite recursive calls.
// See http://choly.ca/post/go-json-marshalling/
type rawRun Run

// mongoRun is the representation of the Run that we store in mongodb.
type mongoRun struct {
	rawRun

	// Bundle is stored in mongo as a string because it has fields that are prefixed with a $, such as $id and $comment.
	// It overrides Run.Bundle.
	Bundle BundleDocument `json:"bundle"`
}

// MarshalJSON converts the run to its storage representation in mongo.
func (r Run) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(mongoRun{
		rawRun: rawRun(r),
		Bundle: BundleDocument(r.Bundle),
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling Run into its storage representation: %w", err)
	}
	return data, nil
}

// UnmarshalJSON converts the run to its storage representation in mongo.
func (r *Run) UnmarshalJSON(data []byte) error {
	var mr mongoRun
	if err := json.Unmarshal(data, &mr); err != nil {
		return fmt.Errorf("error unmarshaling Run from its storage representation: %w", err)
	}

	mr.rawRun.Bundle = bundle.Bundle(mr.Bundle)
	*r = Run(mr.rawRun)
	return nil
}

func (r Run) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"_id": r.ID}
}

// NewRun creates a run with default values initialized.
func NewRun(namespace string, installation string) Run {
	return Run{
		SchemaVersion: DefaultInstallationSchemaVersion,
		ID:            cnab.NewULID(),
		Revision:      cnab.NewULID(),
		Created:       time.Now(),
		Modified:      time.Now(),
		Namespace:     namespace,
		Installation:  installation,
		Parameters:    NewInternalParameterSet(namespace, installation),
	}
}

// ShouldRecord the current run in the Installation history.
// Runs are only recorded for actions that modify the bundle resources,
// or for stateful actions. Stateless actions do not require an existing
// installation or credentials, and are for actions such as documentation, dry-run, etc.
func (r Run) ShouldRecord() bool {
	// Assume all actions modify bundle resources, and should be recorded.
	stateful := true
	modifies := true
	hasOutput := false

	if action, err := r.Bundle.GetAction(r.Action); err == nil {
		modifies = action.Modifies
		stateful = !action.Stateless
	}

	bun := cnab.ExtendedBundle{Bundle: r.Bundle}
	for _, outputDef := range r.Bundle.Outputs {
		if outputDef.AppliesTo(r.Action) && !bun.IsInternalOutput(outputDef.Definition) {
			hasOutput = true
			break
		}
	}

	return modifies || stateful || hasOutput
}

// ToCNAB associated with the Run.
func (r Run) ToCNAB() cnab.Claim {
	return cnab.Claim{
		// CNAB doesn't have the concept of namespace, so we smoosh them together to make a unique name
		SchemaVersion:   cnab.ClaimSchemaVersion(),
		ID:              r.ID,
		Installation:    r.Namespace + "/" + r.Installation,
		Revision:        r.Revision,
		Created:         r.Created,
		Action:          r.Action,
		Bundle:          r.Bundle,
		BundleReference: r.BundleReference,
		Parameters:      r.TypedParameterValues(),
		Custom:          r.Custom,
	}
}

// TypedParameterValues returns parameters values that have been converted to
// its typed value based on its bundle definition.
func (r Run) TypedParameterValues() map[string]interface{} {
	bun := cnab.NewBundle(r.Bundle)
	value := make(map[string]interface{})

	for _, param := range r.Parameters.Parameters {
		v, err := bun.ConvertParameterValue(param.Name, param.ResolvedValue)
		if err != nil {
			value[param.Name] = param.ResolvedValue
			continue
		}
		def, ok := bun.Definitions[param.Name]
		if !ok {
			value[param.Name] = v
			continue
		}
		if bun.IsFileType(def) && v == "" {
			v = nil
		}

		value[param.Name] = v
	}

	return value

}

// SetParametersDigest records the hash of the resolved parameters, so we can
// quickly tell if the parameters between runs were different without
// re-resolving them.
func (r *Run) SetParametersDigest() error {
	// Calculate a hash of the resolved parameters
	paramB, err := json.Marshal(r.Parameters.Parameters)
	if err != nil {
		r.ParametersDigest = ""
		return fmt.Errorf("error calculating the digest of the run parameters: %w", err)
	}

	r.ParametersDigest = fmt.Sprintf("sha256:%x", sha256.Sum256(paramB))
	return nil
}

// SetCredentialsDigest records the hash of the resolved credentials, so we can
// quickly tell if the parameters between runs were different without
// re-resolving them.
func (r *Run) SetCredentialsDigest() error {
	// Calculate a hash of the resolved credentials
	credB, err := json.Marshal(r.Credentials.Credentials)
	if err != nil {
		r.CredentialsDigest = ""
		return fmt.Errorf("error calculating the digest of the run credentials: %w", err)
	}

	r.CredentialsDigest = fmt.Sprintf("sha256:%x", sha256.Sum256(credB))
	return nil
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
		SchemaVersion:  DefaultInstallationSchemaVersion,
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
