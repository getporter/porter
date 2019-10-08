package cnabprovider

import (
	"encoding/base64"
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/cnab-go/claim"
	"github.com/pkg/errors"
)

// loadParameters accepts a set of string overrides and combines that with the default parameters to create
// a full set of parameters.
func (d *Runtime) loadParameters(claim *claim.Claim, rawOverrides map[string]string, action string) (map[string]interface{}, error) {
	overrides := make(map[string]interface{}, len(rawOverrides))
	bun := claim.Bundle

	for key, rawValue := range rawOverrides {
		param, ok := bun.Parameters[key]
		if !ok {
			return nil, fmt.Errorf("parameter %s not defined in bundle", key)
		}

		def, ok := bun.Definitions[param.Definition]
		if !ok {
			return nil, fmt.Errorf("definition %s not defined in bundle", param.Definition)
		}

		unconverted, err := d.getUnconvertedValueFromRaw(def, key, rawValue)
		if err != nil {
			return nil, err
		}

		value, err := def.ConvertValue(unconverted)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to convert parameter's %s value %s to the destination parameter type %s", key, rawValue, def.Type)
		}

		// If this parameter applies to the current action, set the override accordingly
		if param.AppliesTo(action) {
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
	for name, param := range bun.Parameters {
		if param.Required {
			if _, exists := rawOverrides[name]; !exists {
				if !param.AppliesTo(action) {
					overrides[name] = claim.Parameters[name]
				}
			}
		}
	}

	return bundle.ValuesOrDefaults(overrides, bun)
}

func (d *Runtime) getUnconvertedValueFromRaw(def *definition.Schema, key, rawValue string) (string, error) {
	// the parameter value (via rawValue) may represent a file on the local filesystem
	if def.Type == "string" && def.ContentEncoding == "base64" {
		if _, err := d.FileSystem.Stat(rawValue); err == nil {
			bytes, err := d.FileSystem.ReadFile(rawValue)
			if err != nil {
				return "", errors.Wrapf(err, "unable to read file parameter %s", key)
			}
			return base64.StdEncoding.EncodeToString(bytes), nil
		}
	}
	return rawValue, nil
}
