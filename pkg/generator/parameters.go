package generator

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
)

// GenerateParametersOptions are the options to generate a Parameter Set
type GenerateParametersOptions struct {
	GenerateOptions

	// Bundle to generate parameters from
	Bundle cnab.ExtendedBundle
}

// GenerateParameters will generate a parameter set based on the given options
func (opts *GenerateParametersOptions) GenerateParameters() (storage.ParameterSet, error) {
	if opts.Name == "" {
		return storage.ParameterSet{}, errors.New("parameter set name is required")
	}
	generator := genSurvey
	if opts.Silent {
		generator = genEmptySet
	}
	pset, err := opts.genParameterSet(generator)
	if err != nil {
		return storage.ParameterSet{}, err
	}

	pset.Labels = opts.Labels
	return pset, nil
}

func (opts *GenerateParametersOptions) genParameterSet(fn generator) (storage.ParameterSet, error) {
	pset := storage.NewParameterSet(opts.Namespace, opts.Name)

	if strings.ContainsAny(opts.Name, "./\\") {
		return pset, fmt.Errorf("parameter set name '%s' cannot contain the following characters: './\\'", opts.Name)
	}

	var parameterNames []string
	for name := range opts.Bundle.Parameters {
		parameterNames = append(parameterNames, name)
	}

	sort.Strings(parameterNames)

	for _, name := range parameterNames {
		if opts.Bundle.IsInternalParameter(name) {
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
