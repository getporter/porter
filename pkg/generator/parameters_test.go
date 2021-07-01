package generator

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/parameters"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadParametersName(t *testing.T) {
	name := "this.hasadot"
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
	}

	pset, err := opts.GenerateParameters()
	require.Error(t, err, "bad name should have resulted in an error")
	require.Empty(t, pset, "parameter set should have been empty")
	require.EqualError(t, err, fmt.Sprintf("parameter set name '%s' cannot contain the following characters: './\\'", name))
}

func TestGoodParametersName(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
		Bundle: bundle.Bundle{
			Parameters: map[string]bundle.Parameter{
				"one": {
					Definition: "one",
				},
				"two": {
					Definition: "two",
				},
				"three": {
					Definition: "three",
				},
			},
		},
	}

	pset, err := opts.GenerateParameters()
	require.NoError(t, err, "name should NOT have resulted in an error")
	require.Equal(t, 3, len(pset.Parameters), "should have had 3 entries")
}

func TestNoParameters(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
	}
	pset, err := opts.GenerateParameters()
	require.NoError(t, err, "no parameters should have generated an empty parameter set")
	require.NotEmpty(t, pset, "empty parameters should still return an empty parameter set")
}

func TestEmptyParameters(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
		Bundle: bundle.Bundle{},
	}
	pset, err := opts.GenerateParameters()
	require.NoError(t, err, "empty parameters should have generated an empty parameter set")
	require.NotEmpty(t, pset, "empty parameters should still return an empty parameter set")
}

func TestNoParametersName(t *testing.T) {
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Silent: true,
		},
	}
	pset, err := opts.GenerateParameters()
	require.Error(t, err, "expected an error because name is required")
	require.Empty(t, pset, "parameter set should have been empty")
}

func TestSkipParameters(t *testing.T) {
	name := "skip-params"
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
		Bundle: bundle.Bundle{
			Definitions: definition.Definitions{
				"porter-debug": &definition.Schema{
					Comment: parameters.PorterInternal,
				},
			},
			Parameters: map[string]bundle.Parameter{
				"porter-debug": {
					Definition: "porter-debug",
				},
			},
		},
	}

	pset, err := opts.GenerateParameters()
	require.NoError(t, err, "parameters generation should not have resulted in an error")
	assert.Equal(t, "skip-params", pset.Name, "Name was not set")
	require.Empty(t, pset.Parameters, "parameter set should have empty parameters section")
}

func TestDefaultParameters(t *testing.T) {
	firstParamName := "test"
	secondParamName := "test2"
	defaultVal := "example"

	bun := bundle.Bundle{
		Definitions: definition.Definitions{
			firstParamName: &definition.Schema{
				Default: defaultVal,
			},
			secondParamName: &definition.Schema{},
		},
		Parameters: map[string]bundle.Parameter{
			firstParamName: {
				Definition: firstParamName,
			},
			secondParamName: {
				Definition: secondParamName,
			},
		},
	}

	val, err := getDefaultParamValue(bun, firstParamName)
	assert.NoError(t, err, "valid default value for parameter should not give error")
	assert.Equal(t, val, defaultVal)

	val, err = getDefaultParamValue(bun, secondParamName)
	assert.NoError(t, err, "valid default value for parameter should not give error")
	assert.Equal(t, val, nil)
}

func TestMalformedDefaultParameter(t *testing.T) {
	firstParamName := "test"

	bun := bundle.Bundle{
		Definitions: definition.Definitions{},
		Parameters: map[string]bundle.Parameter{
			firstParamName: {
				Definition: firstParamName,
			},
		},
	}

	_, err := getDefaultParamValue(bun, firstParamName)
	assert.NotNil(t, err, "should give error when bundle has no parameter definitions")

	bun2 := bundle.Bundle{
		Definitions: definition.Definitions{
			firstParamName: nil,
		},
		Parameters: map[string]bundle.Parameter{
			firstParamName: {
				Definition: firstParamName,
			},
		},
	}

	_, err = getDefaultParamValue(bun2, firstParamName)
	assert.NotNil(t, err, "parameter with missing definition should give error when fetching it's default value")
}
