package cnabprovider

import (
	"encoding/base64"
	"fmt"

	"github.com/cnabio/cnab-go/valuesource"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/pkg/errors"
)

// loadParameters accepts a set of parameter overrides as well as parameter set
// files and combines both with the default parameters to create a full set
// of parameters.
func (r *Runtime) loadParameters(bun bundle.Bundle, args ActionArguments) (map[string]interface{}, error) {
	mergedParams := make(valuesource.Set, len(args.Params))

	// Apply user supplied parameter overrides last
	for key, rawValue := range args.Params {
		param, ok := bun.Parameters[key]
		if !ok {
			return nil, fmt.Errorf("parameter %s not defined in bundle", key)
		}

		def, ok := bun.Definitions[param.Definition]
		if !ok {
			return nil, fmt.Errorf("definition %s not defined in bundle", param.Definition)
		}

		// Apply porter specific conversions, like retrieving file contents
		value, err := r.getUnconvertedValueFromRaw(def, key, rawValue)
		if err != nil {
			return nil, err
		}

		mergedParams[key] = value
	}

	// Now convert all parameters which are currently strings into the
	// proper type for the parameter, e.g. "false" -> false
	typedParams := make(map[string]interface{}, len(mergedParams))
	for key, unconverted := range mergedParams {
		param, ok := bun.Parameters[key]
		if !ok {
			return nil, fmt.Errorf("parameter %s not defined in bundle", key)
		}

		def, ok := bun.Definitions[param.Definition]
		if !ok {
			return nil, fmt.Errorf("definition %s not defined in bundle", param.Definition)
		}

		value, err := def.ConvertValue(unconverted)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to convert parameter's %s value %s to the destination parameter type %s", key, unconverted, def.Type)
		}

		typedParams[key] = value
	}

	return bundle.ValuesOrDefaults(typedParams, &bun, args.Action)
}

func (r *Runtime) getUnconvertedValueFromRaw(def *definition.Schema, key, rawValue string) (string, error) {
	// the parameter value (via rawValue) may represent a file on the local filesystem
	if def.Type == "string" && def.ContentEncoding == "base64" {
		if _, err := r.FileSystem.Stat(rawValue); err == nil {
			bytes, err := r.FileSystem.ReadFile(rawValue)
			if err != nil {
				return "", errors.Wrapf(err, "unable to read file parameter %s", key)
			}
			return base64.StdEncoding.EncodeToString(bytes), nil
		}
	}
	return rawValue, nil
}
