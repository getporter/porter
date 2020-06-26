package cnabprovider

import (
	"encoding/json"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func Test_loadParameters_paramNotDefined(t *testing.T) {
	d := NewTestRuntime(t)

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
	d := NewTestRuntime(t)

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

func Test_loadParameters_applyTo(t *testing.T) {
	d := NewTestRuntime(t)

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
	require.Equal(t, nil, params["true"], "expected param 'true' to be nil as it does not apply")
}

func Test_loadParameters_applyToBundleDefaults(t *testing.T) {
	d := NewTestRuntime(t)

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

	require.Equal(t, nil, params["foo"], "expected param 'foo' to be nil, regardless of the bundle default, as it does not apply")
}

func Test_loadParameters_requiredButDoesNotApply(t *testing.T) {
	d := NewTestRuntime(t)

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

	require.Equal(t, nil, params["foo"], "expected param 'foo' to be nil, regardless of claim value, as it does not apply")
}

func Test_loadParameters_fileParameter(t *testing.T) {
	d := NewTestRuntime(t)

	d.TestConfig.TestContext.AddTestFile("testdata/file-param", "/path/to/file")

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
			"foo": {
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
					true, true, true, false, nil, "",
				},
				{"required, provided, default does not exist, applies to action",
					true, true, false, true, "my-param-value", "",
				},
				{"required, provided, default does not exist, does not apply to action",
					true, true, false, false, nil, "",
				},
				// As of writing, bundle.ValuesOrDefaults in cnab-go requires a specific override
				// be provided if applicable to an action.
				// Otherwise, it errors out and does not look up/use default in this case.
				{"required, not provided, default exists, applies to action",
					true, false, true, true, nil, "invalid parameters: parameter \"my-param\" is required",
				},
				{"required, not provided, default exists, does not apply to action",
					true, false, true, false, nil, "",
				},
				{"required, not provided, default does not exist, applies to action",
					true, false, false, true, nil, "invalid parameters: parameter \"my-param\" is required",
				},
				{"required, not provided, default does not exist, does not apply to action",
					true, false, false, false, nil, "",
				},
				{"not required, provided, default exists, applies to action",
					false, true, true, true, "my-param-value", "",
				},
				{"not required, provided, default exists, does not apply to action",
					false, true, true, false, nil, "",
				},
				{"not required, provided, default does not exist, applies to action",
					false, true, false, true, "my-param-value", "",
				},
				{"not required, provided, default does not exist, does not apply to action",
					false, true, false, false, nil, "",
				},
				{"not required, not provided, default exists, applies to action",
					false, false, true, true, "my-param-default", "",
				},
				{"not required, not provided, default exists, does not apply to action",
					false, false, true, false, nil, "",
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
					d := NewTestRuntime(t)

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
							"my-param": {
								Definition: "my-param",
								Required:   tc.required,
								Destination: &bundle.Location{
									EnvironmentVariable: "MY_PARAM",
								},
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
						Claim: "test",
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

						err = d.FileSystem.WriteFile("bundle.json", bytes, 0644)
						require.NoError(t, err)

						args.BundlePath = "bundle.json"
					} else {
						// For all other actions, a claim is expected to exist
						// so we create one here and add the bundle to the claim
						claim, err := claim.New("test")
						require.NoError(t, err)

						claim.Bundle = bun
						d.claims.Save(*claim)
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
							updatedClaim, err := d.claims.Read("test")
							require.NoError(t, err)
							require.Equal(t, tc.expectedVal, updatedClaim.Parameters["my-param"])
						}
					}
				})
			}
		})
	}
}
