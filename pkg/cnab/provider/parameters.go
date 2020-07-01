package cnabprovider

import (
	"encoding/base64"
	"fmt"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
)

// loadParameters accepts a set of parameter overrides and combines them
// with the default parameters to create a full set of parameters.
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
	}

	return bundle.ValuesOrDefaults(overrides, bun, action)
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
