package cnabprovider

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/config"
	instancestorage "get.porter.sh/porter/pkg/instance-storage"
	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/cnab-go/claim"
	"github.com/stretchr/testify/require"
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

	overrides := map[string]string{}

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

	require.Equal(t, "foo-claim-value", params["foo"], "expected param 'foo' to be the previous value from the claim")
}

func Test_loadParameters_zeroValues(t *testing.T) {
	var emptyStruct struct{}

	testcases := []struct {
		paramType   string
		expectedVal interface{}
	}{
		{"integer", 0},
		{"number", 0},
		{"string", ""},
		{"boolean", false},
		{"array", []interface{}{}},
		{"object", emptyStruct},
	}

	for _, tc := range testcases {
		t.Run(tc.paramType, func(t *testing.T) {
			c := config.NewTestConfig(t)
			instanceStorage := instancestorage.NewPluggableInstanceStorage(c.Config)
			d := NewRuntime(c.Config, instanceStorage)

			claim, err := claim.New("test")
			require.NoError(t, err)

			claim.Bundle = &bundle.Bundle{
				Definitions: definition.Definitions{
					"foo": &definition.Schema{
						Type: tc.paramType,
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

			claim.Parameters = map[string]interface{}{}
			overrides := map[string]string{}

			params, err := d.loadParameters(claim, overrides, "action")
			require.NoError(t, err)

			require.Equal(t, tc.expectedVal, params["foo"], "unexpected value for param 'foo")
		})
	}
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

// This is intended to cover the matrix of cases around parameter value resolution.
// It exercises the matrix for all supported actions.
func Test_Paramapalooza(t *testing.T) {
	actions := []string{"install", "upgrade", "uninstall", "invoke"}
	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			testcases := []struct {
				name            string
				required        bool
				provided        bool
				defaultExists   bool
				appliesToAction bool
				expectedVal     interface{}
				expectedErr     string
			}{
				// Are you ready to enter the Matrix?
				{"required, provided, default exists, applies to action",
					true, true, true, true, "my-param-value", "",
				},
				{"required, provided, default exists, does not apply to action",
					true, true, true, false, "my-param-value", "",
				},
				{"required, provided, default does not exist, applies to action",
					true, true, false, true, "my-param-value", "",
				},
				{"required, provided, default does not exist, does not apply to action",
					true, true, false, false, "my-param-value", "",
				},
				// As of writing, bundle.ValuesOrDefaults in cnab-go requires a specific override
				// be provided if applicable to an action.
				// Otherwise, it errors out and does not look up/use default in this case.
				{"required, not provided, default exists, applies to action",
					true, false, true, true, nil, "invalid parameters: parameter \"my-param\" is required",
				},
				{"required, not provided, default exists, does not apply to action",
					true, false, true, false, "my-param-default", "",
				},
				{"required, not provided, default does not exist, applies to action",
					true, false, false, true, nil, "invalid parameters: parameter \"my-param\" is required",
				},
				{"required, not provided, default does not exist, does not apply to action",
					true, false, false, false, "", "",
				},
				{"not required, provided, default exists, applies to action",
					false, true, true, true, "my-param-value", "",
				},
				{"not required, provided, default exists, does not apply to action",
					false, true, true, false, "my-param-value", "",
				},
				{"not required, provided, default does not exist, applies to action",
					false, true, false, true, "my-param-value", "",
				},
				{"not required, provided, default does not exist, does not apply to action",
					false, true, false, false, "my-param-value", "",
				},
				{"not required, not provided, default exists, applies to action",
					false, false, true, true, "my-param-default", "",
				},
				{"not required, not provided, default exists, does not apply to action",
					false, false, true, false, "my-param-default", "",
				},
				{"not required, not provided, default does not exist, applies to action",
					false, false, false, true, nil, "",
				},
				{"not required, not provided, default does not exist, does not apply to action",
					false, false, false, false, nil, "",
				},
			}

			for _, tc := range testcases {
				t.Run(tc.name, func(t *testing.T) {
					c := config.NewTestConfig(t)
					instanceStorage := instancestorage.NewTestInstanceStorageProvider()
					d := NewRuntime(c.Config, instanceStorage)

					bun := &bundle.Bundle{
						Name:          "mybuns",
						Version:       "v1.0.0",
						SchemaVersion: "v1.0.0",
						Actions: map[string]bundle.Action{
							"zombies": {
								Modifies: true,
							},
						},
						InvocationImages: []bundle.InvocationImage{
							{
								BaseImage: bundle.BaseImage{
									Image:     "mybuns:latest",
									ImageType: "docker",
								},
							},
						},
						Definitions: definition.Definitions{
							"my-param": &definition.Schema{
								Type: "string",
							},
						},
						Parameters: map[string]bundle.Parameter{
							"my-param": bundle.Parameter{
								Definition: "my-param",
								Required:   tc.required,
							},
						},
					}

					if tc.defaultExists {
						bun.Definitions["my-param"].Default = "my-param-default"
					}

					if !tc.appliesToAction {
						param := bun.Parameters["my-param"]
						param.ApplyTo = []string{"non-applicable-action"}
						bun.Parameters["my-param"] = param
					}

					args := ActionArguments{
						Claim:    "test",
						Insecure: true,
						Driver:   "debug",
					}
					// If param is provided (via --param/--param-file)
					// it will be attached to args
					if tc.provided {
						args.Params = map[string]string{
							"my-param": "my-param-value",
						}
					}

					// If action is install, no claim is expected to exist
					// so we write a bundle and pull it in via ActionArguments
					if action == "install" {
						bytes, err := json.Marshal(bun)
						require.NoError(t, err)

						// We currently need to read/write from the same file on disk
						// as cnab-go's bundle loader still makes raw os calls for loading a bundle
						err = ioutil.WriteFile("testdata/bundle.json", bytes, 0644)
						require.NoError(t, err)

						args.BundlePath = "testdata/bundle.json"
					} else {
						// For all other actions, a claim is expected to exist
						// so we create one here and add the bundle to the claim
						claim, err := claim.New("test")
						require.NoError(t, err)

						claim.Bundle = bun
						d.instanceStorage.Store(*claim)
					}

					var err error
					switch action {
					case "install":
						err = d.Install(args)
					case "upgrade":
						err = d.Upgrade(args)
					case "invoke":
						err = d.Invoke("zombies", args)
					case "uninstall":
						err = d.Uninstall(args)
					}

					if tc.expectedErr != "" {
						require.EqualError(t, err, tc.expectedErr)
					} else {
						require.NoError(t, err)

						if action != "uninstall" {
							// Verify the updated param value on the generated claim
							updatedClaim, err := d.instanceStorage.Read("test")
							require.NoError(t, err)
							require.Equal(t, tc.expectedVal, updatedClaim.Parameters["my-param"])
						}
					}
				})
			}
		})
	}
}
