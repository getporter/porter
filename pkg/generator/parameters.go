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

	// Parameters from the bundle
	Parameters map[string]bundle.Parameter
}

// GenerateParameters will generate a parameter set based on the given options
func GenerateParameters(opts GenerateParametersOptions) (*parameters.ParameterSet, error) {
	if opts.Name == "" {
		return nil, errors.New("parameter set name is required")
	}
	generator := genSurvey
	if opts.Silent {
		generator = genEmptySet
	}
	pset, err := genParameterSet(opts.Name, opts.Parameters, generator)
	if err != nil {
		return nil, err
	}
	return &pset, nil
}

func genParameterSet(name string, params map[string]bundle.Parameter, fn generator) (parameters.ParameterSet, error) {
	pset := parameters.ParameterSet{
		Name: name,
	}
	pset.Parameters = []valuesource.Strategy{}

	if strings.ContainsAny(name, "./\\") {
		return pset, fmt.Errorf("parameter set name '%s' cannot contain the following characters: './\\'", name)
	}

	var parameterNames []string
	for name := range params {
		parameterNames = append(parameterNames, name)
	}

	sort.Strings(parameterNames)

	for _, name := range parameterNames {
		// Skip if parameter is "porter-debug"
		if name == "porter-debug" {
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
