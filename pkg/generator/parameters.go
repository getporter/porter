package generator

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/parameters"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/valuesource"
)

// GenerateParametersOptions are the options to generate a Parameter Set
type GenerateParametersOptions struct {
	GenerateOptions

	// Bundle to generate parameters from
	Bundle bundle.Bundle
}

// GenerateParameters will generate a parameter set based on the given options
func (opts *GenerateParametersOptions) GenerateParameters() (*parameters.ParameterSet, error) {
	if opts.Name == "" {
		return nil, errors.New("parameter set name is required")
	}
	generator := genSurvey
	if opts.Silent {
		generator = genEmptySet
	}
	pset, err := opts.genParameterSet(generator)
	if err != nil {
		return nil, err
	}
	return &pset, nil
}

func (opts *GenerateParametersOptions) genParameterSet(fn generator) (parameters.ParameterSet, error) {
	pset := parameters.ParameterSet{
		Name: opts.Name,
	}
	pset.Parameters = []valuesource.Strategy{}

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
		c, err := fn(name, surveyParameters)
		if err != nil {
			return pset, err
		}
		pset.Parameters = append(pset.Parameters, c)
	}

	return pset, nil
}
