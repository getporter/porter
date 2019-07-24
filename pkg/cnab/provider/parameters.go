package cnabprovider

import (
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/claim"
	"github.com/pkg/errors"
)

// loadParameters accepts a set of string overrides and combines that with the default parameters to create
// a full set of parameters.
func (d *Duffle) loadParameters(claim *claim.Claim, rawOverrides map[string]string, action string) (map[string]interface{}, error) {
	currentVals := claim.Parameters
	overrides := make(map[string]interface{}, len(rawOverrides))
	bun := claim.Bundle

	for key, rawValue := range rawOverrides {
		param, ok := bun.Parameters.Fields[key]
		if !ok {
			return nil, fmt.Errorf("parameter %s not defined in bundle", key)
		}

		def, ok := bun.Definitions[param.Definition]
		if !ok {
			return nil, fmt.Errorf("definition %s not defined in bundle", param.Definition)
		}

		value, err := def.ConvertValue(rawValue)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to convert parameter's %s value %s to the destination parameter type %s", key, rawValue, def.Type)
		}

		// If this parameter applies to the current action, set the override accordingly
		if appliesToAction(action, param) {
			overrides[key] = value
		} else {
			// Otherwise, set to current parameter value on the claim, if exists
			if _, exists := claim.Parameters[key]; exists {
				overrides[key] = claim.Parameters[key]
			}
			if d.Debug {
				fmt.Fprintf(d.Err,
					"override supplied for '%s', but this parameter is not configured to apply for action '%s'; skipping\n",
					key, action)
			}
		}
	}

	// rawOverrides may supply no entry for a parameter designated as required
	// *but* should not apply to this action.
	// When this occurs, we set an override to the current value in the claim such that
	// bundle.ValuesOrDefaults does not return an error
	// TODO: will have to refactor loop once required is back as a bool on the ParameterDefinition
	for _, required := range bun.Parameters.Required {
		if _, exists := rawOverrides[required]; !exists {
			if !appliesToAction(action, bun.Parameters.Fields[required]) {
				overrides[required] = claim.Parameters[required]
			}
		}
	}

	return bundle.ValuesOrDefaults(overrides, currentVals, bun)
}

// TODO: pilfered from cnab-go.  PR to export func?
func appliesToAction(action string, parameter bundle.ParameterDefinition) bool {
	if len(parameter.ApplyTo) == 0 {
		return true
	}
	for _, act := range parameter.ApplyTo {
		if action == act {
			return true
		}
	}
	return false
}
