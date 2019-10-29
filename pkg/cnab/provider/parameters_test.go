package cnabprovider

import (
	"testing"

	instancestorage "github.com/deislabs/porter/pkg/instance-storage"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/stretchr/testify/require"
	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/porter/pkg/config"
)

func Test_loadParameters_paramNotDefined(t *testing.T) {
	c := config.NewTestConfig(t)
	instanceStorage := instancestorage.NewPluggableInstanceStorage(c.Config)
	d := NewRuntime(c.Config, instanceStorage)

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Parameters: map[string]bundle.Parameter{},
	}

	overrides := map[string]string{
		"foo": "bar",
	}

	_, err = d.loadParameters(claim, overrides, "action")
	require.EqualError(t, err, "parameter foo not defined in bundle")
}

func Test_loadParameters_definitionNotDefined(t *testing.T) {
	c := config.NewTestConfig(t)
	instanceStorage := instancestorage.NewPluggableInstanceStorage(c.Config)
	d := NewRuntime(c.Config, instanceStorage)

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Parameters: map[string]bundle.Parameter{
			"foo": bundle.Parameter{
				Definition: "foo",
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
	instanceStorage := instancestorage.NewPluggableInstanceStorage(c.Config)
	d := NewRuntime(c.Config, instanceStorage)

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
		Parameters: map[string]bundle.Parameter{
			"foo": bundle.Parameter{
				Definition: "foo",
				ApplyTo: []string{
					"action",
				},
			},
			"bar": bundle.Parameter{
				Definition: "bar",
			},
			"true": bundle.Parameter{
				Definition: "true",
				ApplyTo: []string{
					"different-action",
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
	instanceStorage := instancestorage.NewPluggableInstanceStorage(c.Config)
	d := NewRuntime(c.Config, instanceStorage)

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type:    "string",
				Default: "foo-default",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"foo": bundle.Parameter{
				Definition: "foo",
				ApplyTo: []string{
					"different-action",
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
	instanceStorage := instancestorage.NewPluggableInstanceStorage(c.Config)
	d := NewRuntime(c.Config, instanceStorage)

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type: "string",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"foo": bundle.Parameter{
				Definition: "foo",
				ApplyTo: []string{
					"different-action",
				},
				Required: true,
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

func Test_loadParameters_fileParameter(t *testing.T) {
	c := config.NewTestConfig(t)
	instanceStorage := instancestorage.NewPluggableInstanceStorage(c.Config)
	d := NewRuntime(c.Config, instanceStorage)

	c.TestContext.AddTestFile("testdata/file-param", "/path/to/file")

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type:            "string",
				ContentEncoding: "base64",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"foo": bundle.Parameter{
				Definition: "foo",
				Required:   true,
				Destination: &bundle.Location{
					Path: "/tmp/foo",
				},
			},
		},
	}

	overrides := map[string]string{
		"foo": "/path/to/file",
	}

	params, err := d.loadParameters(claim, overrides, "action")
	require.NoError(t, err)

	require.Equal(t, "SGVsbG8gV29ybGQh", params["foo"], "expected param 'foo' to be the base64-encoded file contents")
}
