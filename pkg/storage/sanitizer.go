package storage

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/pkg/errors"
)

// Sanitizer identifies sensitive data in a database record, and replaces it with
// a reference to a secret created by the service in an external secret store.
type Sanitizer struct {
	parameter ParameterSetProvider
	secrets   secrets.Store
}

// NewSanitizer creates a new service for sanitizing sensitive data and save them
// to a secret store.
func NewSanitizer(parameterstore ParameterSetProvider, secretstore secrets.Store) *Sanitizer {
	return &Sanitizer{
		parameter: parameterstore,
		secrets:   secretstore,
	}
}

// CleanRawParameters clears out sensitive data in raw parameter values before
// transform the raw value into secret strategies.
// The id argument is used to associate the reference key with the corresponding
// run or installation record in porter's database.
func (s *Sanitizer) CleanRawParameters(ctx context.Context, params map[string]interface{}, bun cnab.ExtendedBundle, id string) ([]secrets.Strategy, error) {
	strategies := make([]secrets.Strategy, 0, len(params))
	for name, value := range params {
		stringVal, err := bun.WriteParameterToString(name, value)
		if err != nil {
			return nil, err
		}
		strategy := ValueStrategy(name, stringVal)
		strategies = append(strategies, strategy)
	}

	strategies, err := s.CleanParameters(ctx, strategies, bun, id)
	if err != nil {
		return nil, err
	}

	return strategies, nil

}

// CleanParameters clears out sensitive data in strategized parameter data and return
// sanitized value after saving sensitive datat to secrets store.
// The id argument is used to associate the reference key with the corresponding
// run or installation record in porter's database.
func (s *Sanitizer) CleanParameters(ctx context.Context, params []secrets.Strategy, bun cnab.ExtendedBundle, id string) ([]secrets.Strategy, error) {
	strategies := make([]secrets.Strategy, 0, len(params))
	for _, param := range params {
		if param.Source.Key == secrets.SourceSecret {
			strategies = append(strategies, param)
			continue
		}
		strategy := ValueStrategy(param.Name, param.Value)
		if bun.IsSensitiveParameter(param.Name) {
			cleaned := sanitizedParam(strategy, id)
			err := s.secrets.Create(ctx, cleaned.Source.Key, cleaned.Source.Value, cleaned.Value)
			if err != nil {
				return nil, errors.Wrap(err, "failed to save sensitive param to secrete store")
			}
			strategy = cleaned
		}

		strategies = append(strategies, strategy)
	}

	if len(strategies) == 0 {
		return nil, nil
	}

	return strategies, nil

}

// LinkSensitiveParametersToSecrets creates a reference key for sensitive data
// and replace the sensitive value with the reference key.
// The id argument is used to associate the reference key with the corresponding
// run or installation record in porter's database.
func LinkSensitiveParametersToSecrets(pset ParameterSet, bun cnab.ExtendedBundle, id string) ParameterSet {
	for i, param := range pset.Parameters {
		if !bun.IsSensitiveParameter(param.Name) {
			continue
		}
		pset.Parameters[i] = sanitizedParam(param, id)
	}

	return pset
}

func sanitizedParam(param secrets.Strategy, id string) secrets.Strategy {
	param.Source.Key = secrets.SourceSecret
	param.Source.Value = id + param.Name
	return param
}

// RestoreParameterSet resolves the raw parameter data from a secrets store.
func (s *Sanitizer) RestoreParameterSet(ctx context.Context, pset ParameterSet, bun cnab.ExtendedBundle) (map[string]interface{}, error) {
	params, err := s.parameter.ResolveAll(ctx, pset)
	if err != nil {
		return nil, err
	}

	resolved := make(map[string]interface{})
	for name, value := range params {
		paramValue, err := bun.ConvertParameterValue(name, value)
		if err != nil {
			paramValue = value
		}

		resolved[name] = paramValue

	}
	return resolved, nil

}

// CleanOutput clears data that's defined as sensitive on the bundle definition
// by storing the raw data into a secret store and store it's reference key onto
// the output record.
func (s *Sanitizer) CleanOutput(ctx context.Context, output Output, bun cnab.ExtendedBundle) (Output, error) {
	// Skip outputs not defined in the bundle, e.g. io.cnab.outputs.invocationImageLogs
	_, ok := output.GetSchema(bun)
	if !ok {
		return output, nil
	}

	sensitive, err := bun.IsOutputSensitive(output.Name)
	if err != nil {
		output.Value = nil
		return output, err
	}

	if !sensitive {
		return output, nil

	}

	secretOt := sanitizedOutput(output)

	err = s.secrets.Create(ctx, secrets.SourceSecret, secretOt.Key, string(output.Value))
	if err != nil {
		return secretOt, err
	}

	return secretOt, nil
}

func sanitizedOutput(output Output) Output {
	output.Key = output.RunID + output.Name
	output.Value = nil
	return output

}

// RestoreOutputs retrieves all raw output value and return the restored outputs
// record.
func (s *Sanitizer) RestoreOutputs(ctx context.Context, o Outputs) (Outputs, error) {
	resolved := make([]Output, 0, o.Len())
	for _, ot := range o.Value() {
		r, err := s.RestoreOutput(ctx, ot)
		if err != nil {
			return o, errors.WithMessagef(err, "failed to resolve output %q using key %q", ot.Name, ot.Key)
		}
		resolved = append(resolved, r)
	}

	return NewOutputs(resolved), nil
}

// RestoreOutput retrieves the raw output value and return the restored output
// record.
func (s *Sanitizer) RestoreOutput(ctx context.Context, output Output) (Output, error) {
	if output.Key == "" {
		return output, nil
	}
	resolved, err := s.secrets.Resolve(ctx, secrets.SourceSecret, string(output.Key))
	if err != nil {
		return output, err
	}

	output.Value = []byte(resolved)
	return output, nil
}
