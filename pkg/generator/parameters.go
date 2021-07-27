package generator

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/parameters"
	"github.com/cnabio/cnab-go/bundle"
)

// GenerateParametersOptions are the options to generate a Parameter Set
type GenerateParametersOptions struct {
	GenerateOptions

	// Bundle to generate parameters from
	Bundle bundle.Bundle
}

// GenerateParameters will generate a parameter set based on the given options
func (opts *GenerateParametersOptions) GenerateParameters() (parameters.ParameterSet, error) {
	if opts.Name == "" {
		return parameters.ParameterSet{}, errors.New("parameter set name is required")
	}
	generator := genSurvey
	if opts.Silent {
		generator = genEmptySet
	}
	pset, err := opts.genParameterSet(generator)
	if err != nil {
		return parameters.ParameterSet{}, err
	}
	return pset, nil
}

func (opts *GenerateParametersOptions) genParameterSet(fn generator) (parameters.ParameterSet, error) {
	pset := parameters.NewParameterSet(opts.Name)

	if strings.ContainsAny(opts.Name, "./\\") {
		return pset, fmt.Errorf("parameter set name '%s' cannot contain the following characters: './\\'", opts.Name)
	}

	var parameterNames []string
	for name := range opts.Bundle.Parameters {
		parameterNames = append(parameterNames, name)
	}

	sort.Strings(parameterNames)

	for _, name := range parameterNames {
		if parameters.IsInternal(name, opts.Bundle) {
			continue
		}
		defaultVal, err := getDefaultParamValue(opts.Bundle, name)

		if err != nil {
			return pset, err
		}

		c, err := fn(name, surveyParameters, defaultVal)
		if err != nil {
			return pset, err
		}

		// If any of name or source info is missing, do not include this
		// parameter in the parameter set
		if c.Name != "" && c.Source.Key != "" && c.Source.Value != "" {
			pset.Parameters = append(pset.Parameters, c)
		}
	}

	return pset, nil
}

func getDefaultParamValue(bun bundle.Bundle, name string) (interface{}, error) {
	for p, v := range bun.Parameters {
		if p == name {
			def, ok := bun.Definitions[v.Definition]
			if !ok {
				return "", fmt.Errorf("unable to find definition for parameter %s", name)
			}
			if def == nil {
				return "", fmt.Errorf("parameter definition for %s is empty", name)
			}

			return def.Default, nil
		}
	}
	return "", nil
}
