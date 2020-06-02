package cnabprovider

import (
	"encoding/base64"
	"fmt"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/pkg/errors"
)

// loadParameters accepts a set of parameter overrides as well as parameter set
// files and combines both with the default parameters to create a full set
// of parameters.
func (d *Runtime) loadParameters(claim *claim.Claim, rawOverrides map[string]string, parameterSets []string, action string) (map[string]interface{}, error) {
	overrides := make(map[string]interface{}, len(rawOverrides))
	bun := claim.Bundle

	// Loop through each parameter set file and load the parameter values
	loaded, err := d.loadParameterSets(bun, parameterSets)
	if err != nil {
		return nil, errors.Wrap(err, "unable to process provided parameter sets")
	}

	for key, val := range loaded {
		overrides[key] = val
	}

	// Now give precedence to the raw overrides that came via the CLI
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

// loadParameterSets loads parameter values per their parameter set strategies
func (d *Runtime) loadParameterSets(b *bundle.Bundle, params []string) (valuesource.Set, error) {
	resolvedParameters := valuesource.Set{}
	for _, name := range params {
		pset, err := d.parameters.Read(name)
		if err != nil {
			return nil, err
		}

		rc, err := d.parameters.ResolveAll(pset)
		if err != nil {
			return nil, err
		}

		for k, v := range rc {
			resolvedParameters[k] = v
		}
	}

	return resolvedParameters, nil
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
