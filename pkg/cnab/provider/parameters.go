package cnabprovider

import (
	"encoding/base64"
	"fmt"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"

	"get.porter.sh/porter/pkg/cnab/extensions"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/pkg/errors"
)

// loadParameters accepts a set of parameter overrides as well as parameter set
// files and combines both with the default parameters to create a full set
// of parameters.
func (r *Runtime) loadParameters(bun bundle.Bundle, args ActionArguments) (map[string]interface{}, error) {
	mergedParams := make(secrets.Set, len(args.Params))
	paramSources, err := r.resolveParameterSources(bun, args)
	if err != nil {
		return nil, err
	}

	for key, val := range paramSources {
		mergedParams[key] = val
	}

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
		value, err := r.getUnconvertedValueFromRaw(bun, def, key, rawValue)
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

		if def.Type != nil {
			value, err := def.ConvertValue(unconverted)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to convert parameter's %s value %s to the destination parameter type %s", key, unconverted, def.Type)
			}
			typedParams[key] = value
		} else {
			// bundle dependency parameters can be any type, not sure we have a solid way to do a typed conversion
			typedParams[key] = unconverted
		}

	}

	return bundle.ValuesOrDefaults(typedParams, &bun, args.Action)
}

func (r *Runtime) getUnconvertedValueFromRaw(b bundle.Bundle, def *definition.Schema, key, rawValue string) (string, error) {
	// the parameter value (via rawValue) may represent a file on the local filesystem
	if extensions.IsFileType(b, def) {
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

func (r *Runtime) resolveParameterSources(bun bundle.Bundle, args ActionArguments) (secrets.Set, error) {
	if r.Debug {
		fmt.Fprintln(r.Err, "Resolving parameter sources...")
	}
	parameterSources, required, err := r.Extensions.GetParameterSources()
	if err != nil {
		return nil, err
	}

	if !required {
		if r.Debug {
			fmt.Fprintln(r.Err, "No parameter sources defined!")
		}
		return nil, nil
	}

	values := secrets.Set{}
	for parameterName, parameterSource := range parameterSources {
		if r.Debug {
			fmt.Fprintln(r.Err, "Resolving parameter source", parameterName)
		}
		for _, rawSource := range parameterSource.ListSourcesByPriority() {
			var installation string
			var outputName string
			switch source := rawSource.(type) {
			case extensions.OutputParameterSource:
				installation = args.Installation.Name
				outputName = source.OutputName
			case extensions.DependencyOutputParameterSource:
				// TODO(carolynvs): does this need to take namespace into account
				installation = extensions.BuildPrerequisiteInstallationName(args.Installation.Name, source.Dependency)
				outputName = source.OutputName
			}

			output, err := r.claims.GetLastOutput(args.Installation.Namespace, installation, outputName)
			if err != nil {
				// When we can't find the output, skip it and let the parameter be set another way
				if errors.Is(err, storage.ErrNotFound{}) {
					if r.Debug {
						fmt.Fprintf(r.Err, "No previous output found for %s from %s/%s\n", outputName, args.Installation.Namespace, installation)
					}
					continue
				}
				// Otherwise, something else has happened, perhaps bad data or connectivity problems, we can't ignore it
				return nil, errors.Wrapf(err, "could not set parameter %s from output %s of %s", parameterName, outputName, installation)
			}

			param, ok := bun.Parameters[parameterName]
			if !ok {
				return nil, fmt.Errorf("parameter %s not defined in bundle", parameterName)
			}

			def, ok := bun.Definitions[param.Definition]
			if !ok {
				return nil, fmt.Errorf("definition %s not defined in bundle", param.Definition)
			}

			if extensions.IsFileType(bun, def) {
				values[parameterName] = base64.StdEncoding.EncodeToString(output.Value)
			} else {
				values[parameterName] = string(output.Value)
			}

			if r.Debug {
				fmt.Fprintf(r.Out, "Injected installation %s output %s as parameter %s\n", installation, outputName, parameterName)
			}
		}
	}

	return values, nil
}
