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

		overrides[key] = value
		// If this parameter does not apply to the current action, defer to the claim value, if exists
		if !param.AppliesTo(action) {
			if _, exists := claim.Parameters[key]; exists {
				overrides[key] = claim.Parameters[key]
			}
		}
	}

	// rawOverrides (meaning, user-supplied overrides at time of action invocation)
	// may supply no entry for a parameter designated as required *but* that does not apply to this action.
	//
	// When this occurs, we set an override to either the current value in the claim or the default value.
	// If neither exists, the zero value according to the parameter type will be used.
	// Otherwise, if unset/nil, json validation in bundle.ValuesOrDefaults would return an error
	for name, param := range bun.Parameters {
		def, ok := bun.Definitions[param.Definition]
		if !ok {
			return nil, fmt.Errorf("parameter definition %s not defined in bundle", param.Definition)
		}
		if param.Required {
			if _, exists := rawOverrides[name]; !exists {
				if !param.AppliesTo(action) {
					// First defer to a pre-existing value in the claim
					if claim.Parameters[name] != nil {
						overrides[name] = claim.Parameters[name]
					} else if def.Default != nil {
						// Next defer to a default value
						overrides[name] = def.Default
					} else {
						// Finally, use a zero value if no other option exists
						overrides[name] = getZeroValue(name, def)
					}
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

// getZeroValue returns the zero value for a parameter according to its type
func getZeroValue(name string, def *definition.Schema) interface{} {
	switch def.Type {
	case "integer", "number":
		return 0
	case "string":
		return ""
	case "boolean":
		return false
	case "array":
		return []interface{}{}
	case "object":
		var emptyStruct struct{}
		return emptyStruct
	default:
		return nil
	}
}
