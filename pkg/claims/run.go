package claims

import (
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/schema"
	"github.com/pkg/errors"
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

	// ParameterOverrides are the key/value parameter overrides (taking precedence over
	// parameters specified in a parameter set) specified during the run.
	ParameterOverrides parameters.ParameterSet `json:"parameterOverrides, omitempty" yaml:"parameterOverrides, omitempty", toml:"parameterOverrides, omitempty"`

	// CredentialSets is a list of the credential set names used during the run.
	CredentialSets []string `json:"credentialSets,omitempty" yaml:"credentialSets,omitempty" toml:"credentialSets,omitempty"`

	// ParameterSets is the list of parameter set names used during the run.
	ParameterSets []string `json:"parameterSets,omitempty" yaml:"parameterSets,omitempty" toml:"parameterSets,omitempty"`

	// Parameters is the full set of resolved parameters stored on the claim.
	// This includes internal parameters, resolved parameter sources, values resolved from parameter sets, etc.
	ResolvedParameters map[string]interface{} `json:"-" yaml:"-" toml:"-"`

	// Parameters is the full set of parameters that's being used during the
	// current run.
	// This includes internal parameters, parameter sources, values from parameter sets, etc.
	// Any sensitive data will be sannitized before saving to the database.
	Parameters parameters.ParameterSet `json:"parameters,omitempty" yaml:"parameters,omitempty" toml:"parameters,omitempty"`

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
		SchemaVersion:      SchemaVersion,
		ID:                 cnab.NewULID(),
		Revision:           cnab.NewULID(),
		Created:            time.Now(),
		Namespace:          namespace,
		Installation:       installation,
		ResolvedParameters: make(map[string]interface{}),
		Parameters:         parameters.NewInternalParameterSet(namespace, installation),
	}
}

// PopulateParameters populates the Parameters field with provided parameters.
// It also saves any sensitive value into the provided secret store.
func (r *Run) PopulateParameters(params map[string]interface{}, store secrets.Store) error {
	strategies := make([]secrets.Strategy, 0, len(params))
	bun := cnab.ExtendedBundle{r.Bundle}
	for name, value := range params {

		stringVal, err := bun.WriteParameterToString(name, value)
		if err != nil {
			return err

		}

		strategy := parameters.DefaultStrategy(name, stringVal)

		if bun.IsSensitiveParameter(name) {
			encodedStrategy := r.EncodeSensitiveParameter(strategy)
			err := store.Create(encodedStrategy.Source.Key, encodedStrategy.Source.Value, encodedStrategy.Value)
			if err != nil {
				return errors.Wrap(err, "failed to save sensitive param to secrete store")
			}
			strategy = encodedStrategy
		}

		strategies = append(strategies, strategy)
	}

	if len(strategies) == 0 {
		return nil
	}

	r.Parameters.Parameters = strategies
	return nil

}

// ShouldRecord the current run in the Installation history.
// Runs are only recorded for actions that modify the bundle resources,
// or for stateful actions. Stateless actions do not require an existing
// installation or credentials, and are for actions such as documentation, dry-run, etc.
func (r Run) ShouldRecord() bool {
	// Assume all actions modify bundle resources, and should be recorded.
	stateful := true
	modifies := true

	if action, err := r.Bundle.GetAction(r.Action); err == nil {
		modifies = action.Modifies
		stateful = !action.Stateless
	}

	return modifies || stateful
}

// ToCNAB associated with the Run.
func (r Run) ToCNAB() cnab.Claim {
	return cnab.Claim{
		// CNAB doesn't have the concept of namespace, so we smoosh them together to make a unique name
		SchemaVersion:   CNABSchemaVersion(),
		ID:              r.ID,
		Installation:    r.Namespace + "/" + r.Installation,
		Revision:        r.Revision,
		Created:         r.Created,
		Action:          r.Action,
		Bundle:          r.Bundle,
		BundleReference: r.BundleReference,
		Parameters:      r.ResolvedParameters,
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

// EncodeInternalParameterSet encodes sensitive data in internal parameter sets
// so it will be associated with the current run record.
func (r *Run) EncodeParameterOverrides() {

	bun := cnab.ExtendedBundle{r.Bundle}

	for i, param := range r.ParameterOverrides.Parameters {
		if !bun.IsSensitiveParameter(param.Name) {
			continue
		}
		r.ParameterOverrides.Parameters[i] = r.EncodeSensitiveParameter(param)
	}

}

// EncodeSensitiveParameter returns a new parameter that has the run ID
// associated with the strategy.
func (r Run) EncodeSensitiveParameter(param secrets.Strategy) secrets.Strategy {
	param.Source.Key = secrets.SourceSecret
	param.Source.Value = r.ID + param.Name
	return param
}

// Resolve resolves reference values on a run record.
// Currently, it's resolving sensitive parameter values.
func (r Run) Resolve(resolver parameters.Provider) (Run, error) {
	bun := cnab.ExtendedBundle{r.Bundle}

	parameterOverrides, err := r.ParameterOverrides.Resolve(resolver, bun)
	if err != nil {
		return r, err
	}
	params, err := r.Parameters.Resolve(resolver, bun)
	if err != nil {
		return r, err
	}
	if r.ResolvedParameters == nil {
		r.ResolvedParameters = make(map[string]interface{})
	}
	for key, value := range params {
		r.ResolvedParameters[key] = value
	}
	for key, value := range parameterOverrides {
		r.ResolvedParameters[key] = value
	}

	return r, nil
}
