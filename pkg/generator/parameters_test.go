package generator

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/parameters"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/valuesource"
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

	pset, err := GenerateParameters(opts)
	require.Error(t, err, "bad name should have resulted in an error")
	require.Nil(t, pset, "parameter set should have been empty")
	require.EqualError(t, err, fmt.Sprintf("parameter set name '%s' cannot contain the following characters: './\\'", name))
}

func TestGoodParametersName(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
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
	}

	pset, err := GenerateParameters(opts)
	require.NoError(t, err, "name should NOT have resulted in an error")
	require.NotNil(t, pset, "parameter set should NOT have been empty")
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
	pset, err := GenerateParameters(opts)
	require.NoError(t, err, "no parameters should have generated an empty parameter set")
	require.NotNil(t, pset, "empty parameters should still return an empty parameter set")
}

func TestEmptyParameters(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
		Parameters: map[string]bundle.Parameter{},
	}
	pset, err := GenerateParameters(opts)
	require.NoError(t, err, "empty parameters should have generated an empty parameter set")
	require.NotNil(t, pset, "empty parameters should still return an empty parameter set")
}

func TestNoParametersName(t *testing.T) {
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Silent: true,
		},
	}
	pset, err := GenerateParameters(opts)
	require.Error(t, err, "expected an error because name is required")
	require.Nil(t, pset, "parameter set should have been empty")
}

func TestSkipParameters(t *testing.T) {
	name := "skip-params"
	opts := GenerateParametersOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
		Parameters: map[string]bundle.Parameter{
			"porter-debug": {
				Definition: "porter-debug",
			},
		},
	}

	expected := &parameters.ParameterSet{
		Name:       "skip-params",
		Parameters: []valuesource.Strategy{},
	}

	pset, err := GenerateParameters(opts)
	require.NoError(t, err, "parameters generation should not have resulted in an error")
	require.NotNil(t, pset, "parameter set should not be nil")
	require.Equal(t, expected, pset, "parameter set should have empty parameters section")
}
