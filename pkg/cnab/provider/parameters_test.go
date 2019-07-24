package cnabprovider

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle/definition"

	"github.com/stretchr/testify/require"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/claim"

	"github.com/deislabs/porter/pkg/config"
)

func Test_loadParameters_paramNotDefined(t *testing.T) {
	c := config.NewTestConfig(t)
	d := NewDuffle(c.Config)

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Parameters: &bundle.ParametersDefinition{
			Fields: map[string]bundle.ParameterDefinition{},
		},
	}

	overrides := map[string]string{
		"foo": "bar",
	}

	_, err = d.loadParameters(claim, overrides, "action")
	require.EqualError(t, err, "parameter foo not defined in bundle")
}

func Test_loadParameters_definitionNotDefined(t *testing.T) {
	c := config.NewTestConfig(t)
	d := NewDuffle(c.Config)

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Parameters: &bundle.ParametersDefinition{
			Fields: map[string]bundle.ParameterDefinition{
				"foo": bundle.ParameterDefinition{
					Definition: "foo",
				},
			},
		},
	}

	overrides := map[string]string{
		"foo": "bar",
	}

	_, err = d.loadParameters(claim, overrides, "action")
	require.EqualError(t, err, "definition foo not defined in bundle")
}

func Test_loadParameters_applyToClaimDefaults(t *testing.T) {
	c := config.NewTestConfig(t)
	d := NewDuffle(c.Config)

	claim, err := claim.New("test")
	require.NoError(t, err)

	// Here we set default values, but we expect the corresponding
	// claim values to take precedence when loadParameters is called
	claim.Bundle = &bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type:    "string",
				Default: "default-foo-value",
			},
			"bar": &definition.Schema{
				Type:    "integer",
				Default: "default-bar-value",
			},
			"true": &definition.Schema{
				Type:    "boolean",
				Default: "default-true-value",
			},
		},
		Parameters: &bundle.ParametersDefinition{
			Fields: map[string]bundle.ParameterDefinition{
				"foo": bundle.ParameterDefinition{
					Definition: "foo",
					ApplyTo: []string{
						"action",
					},
				},
				"bar": bundle.ParameterDefinition{
					Definition: "bar",
				},
				"true": bundle.ParameterDefinition{
					Definition: "true",
					ApplyTo: []string{
						"different-action",
					},
				},
			},
		},
	}

	claim.Parameters = map[string]interface{}{
		"foo":  "foo",
		"bar":  123,
		"true": true,
	}

	overrides := map[string]string{
		"foo":  "FOO",
		"bar":  "456",
		"true": "false",
	}

	params, err := d.loadParameters(claim, overrides, "action")
	require.NoError(t, err)

	require.Equal(t, "FOO", params["foo"], "expected param 'foo' to be updated")
	require.Equal(t, 456, params["bar"], "expected param 'bar' to be updated")
	require.Equal(t, true, params["true"], "expected param 'true' to represent the preexisting claim value")
}

func Test_loadParameters_applyToBundleDefaults(t *testing.T) {
	c := config.NewTestConfig(t)
	d := NewDuffle(c.Config)

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type:    "string",
				Default: "foo-default",
			},
		},
		Parameters: &bundle.ParametersDefinition{
			Fields: map[string]bundle.ParameterDefinition{
				"foo": bundle.ParameterDefinition{
					Definition: "foo",
					ApplyTo: []string{
						"different-action",
					},
				},
			},
		},
	}

	claim.Parameters = map[string]interface{}{}

	overrides := map[string]string{
		"foo": "FOO",
	}

	params, err := d.loadParameters(claim, overrides, "action")
	require.NoError(t, err)

	require.Equal(t, "foo-default", params["foo"], "expected param 'foo' to be the bundle default")
}

func Test_loadParameters_requiredButDoesNotApply(t *testing.T) {
	c := config.NewTestConfig(t)
	d := NewDuffle(c.Config)

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type: "string",
			},
		},
		Parameters: &bundle.ParametersDefinition{
			Fields: map[string]bundle.ParameterDefinition{
				"foo": bundle.ParameterDefinition{
					Definition: "foo",
					ApplyTo: []string{
						"different-action",
					},
				},
			},
			Required: []string{
				"foo",
			},
		},
	}

	claim.Parameters = map[string]interface{}{
		"foo": "foo-claim-value",
	}

	overrides := map[string]string{}

	params, err := d.loadParameters(claim, overrides, "action")
	require.NoError(t, err)

	require.Equal(t, "foo-claim-value", params["foo"], "expected param 'foo' to be the bundle default")
}
