package cnabprovider

import (
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_loadParameters_paramNotDefined(t *testing.T) {
	r := NewTestRuntime(t)
	b := bundle.Bundle{
		Parameters: map[string]bundle.Parameter{},
	}

	overrides := map[string]string{
		"foo": "bar",
	}

	args := ActionArguments{
		Action: "action",
		Params: overrides,
	}
	_, err := r.loadParameters(b, args)
	require.EqualError(t, err, "parameter foo not defined in bundle")
}

func Test_loadParameters_definitionNotDefined(t *testing.T) {
	r := NewTestRuntime(t)

	b := bundle.Bundle{
		Parameters: map[string]bundle.Parameter{
			"foo": {
				Definition: "foo",
			},
		},
	}

	overrides := map[string]string{
		"foo": "bar",
	}

	args := ActionArguments{
		Action: "action",
		Params: overrides,
	}
	_, err := r.loadParameters(b, args)
	require.EqualError(t, err, "definition foo not defined in bundle")
}

func Test_loadParameters_applyTo(t *testing.T) {
	r := NewTestRuntime(t)

	// Here we set default values, but we expect the corresponding
	// claim values to take precedence when loadParameters is called
	b := bundle.Bundle{
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
			"foo": {
				Definition: "foo",
				ApplyTo: []string{
					"action",
				},
			},
			"bar": {
				Definition: "bar",
			},
			"true": {
				Definition: "true",
				ApplyTo: []string{
					"different-action",
				},
			},
		},
	}

	overrides := map[string]string{
		"foo":  "FOO",
		"bar":  "456",
		"true": "false",
	}

	args := ActionArguments{
		Action: "action",
		Params: overrides,
	}
	params, err := r.loadParameters(b, args)
	require.NoError(t, err)

	require.Equal(t, "FOO", params["foo"], "expected param 'foo' to be updated")
	require.Equal(t, 456, params["bar"], "expected param 'bar' to be updated")
	require.Equal(t, nil, params["true"], "expected param 'true' to be nil as it does not apply")
}

func Test_loadParameters_applyToBundleDefaults(t *testing.T) {
	r := NewTestRuntime(t)

	b := bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type:    "string",
				Default: "foo-default",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"foo": {
				Definition: "foo",
				ApplyTo: []string{
					"different-action",
				},
			},
		},
	}

	args := ActionArguments{Action: "action"}
	params, err := r.loadParameters(b, args)
	require.NoError(t, err)

	require.Equal(t, nil, params["foo"], "expected param 'foo' to be nil, regardless of the bundle default, as it does not apply")
}

func Test_loadParameters_requiredButDoesNotApply(t *testing.T) {
	r := NewTestRuntime(t)

	b := bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type: "string",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"foo": {
				Definition: "foo",
				ApplyTo: []string{
					"different-action",
				},
				Required: true,
			},
		},
	}

	args := ActionArguments{Action: "action"}
	params, err := r.loadParameters(b, args)
	require.NoError(t, err)

	require.Equal(t, nil, params["foo"], "expected param 'foo' to be nil, regardless of claim value, as it does not apply")
}

func Test_loadParameters_fileParameter(t *testing.T) {
	r := NewTestRuntime(t)

	r.TestConfig.TestContext.AddTestFile("testdata/file-param", "/path/to/file")

	b := bundle.Bundle{
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

	args := ActionArguments{
		Action: "action",
		Params: overrides,
	}
	params, err := r.loadParameters(b, args)
	require.NoError(t, err)

	require.Equal(t, "SGVsbG8gV29ybGQh", params["foo"], "expected param 'foo' to be the base64-encoded file contents")
}

func Test_loadParameters_ParameterSourcePrecedence(t *testing.T) {
	r := NewTestRuntime(t)
	r.TestParameters.AddTestParameters("testdata/paramset.json")
	r.TestParameters.TestSecrets.AddSecret("foo_secret", "foo_set")

	r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
	b, err := r.ProcessBundle("bundle.json")
	require.NoError(t, err, "ProcessBundle failed")

	overrides := map[string]string{
		"foo": "foo_override",
	}

	t.Run("nothing present, use default", func(t *testing.T) {
		args := ActionArguments{
			Installation: "mybun",
			Action:       claim.ActionUpgrade,
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "foo_default", params["foo"],
			"expected param 'foo' to have default value")
	})

	t.Run("only override present", func(t *testing.T) {
		args := ActionArguments{
			Installation: "mybun",
			Action:       claim.ActionUpgrade,
			Params:       overrides,
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "foo_override", params["foo"],
			"expected param 'foo' to have override value")
	})

	t.Run("only parameter source present", func(t *testing.T) {
		tc := r.claims.(*claims.TestClaimProvider)
		c := tc.CreateClaim("mybun", claim.ActionInstall, b, nil)
		cr := tc.CreateResult(c, claim.StatusSucceeded)
		tc.CreateOutput(c, cr, "foo", []byte("foo_source"))

		args := ActionArguments{
			Installation: "mybun",
			Action:       claim.ActionUpgrade,
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "foo_source", params["foo"],
			"expected param 'foo' to have parameter source value")
	})

	t.Run("override > parameter source", func(t *testing.T) {
		tc := r.claims.(*claims.TestClaimProvider)
		c := tc.CreateClaim("mybun", claim.ActionInstall, b, nil)
		cr := tc.CreateResult(c, claim.StatusSucceeded)
		tc.CreateOutput(c, cr, "foo", []byte("foo_source"))

		args := ActionArguments{
			Installation: "mybun",
			Action:       claim.ActionUpgrade,
			Params:       overrides,
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "foo_override", params["foo"],
			"expected param 'foo' to have parameter override value")
	})

	t.Run("merge parameter values", func(t *testing.T) {
		// foo is set by a the user
		// baz is set by a parameter source
		// bar is set by the bundle default
		tc := r.claims.(*claims.TestClaimProvider)
		c := tc.CreateClaim("mybun", claim.ActionInstall, b, nil)
		cr := tc.CreateResult(c, claim.StatusSucceeded)
		tc.CreateOutput(c, cr, "foo", []byte("foo_source"))
		tc.CreateOutput(c, cr, "bar", []byte("bar_source"))
		tc.CreateOutput(c, cr, "baz", []byte("baz_source"))

		args := ActionArguments{
			Installation: "mybun",
			Action:       claim.ActionUpgrade,
			Params:       map[string]string{"foo": "foo_override"},
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "foo_override", params["foo"],
			"expected param 'foo' to have parameter override value")
		assert.Equal(t, "bar_source", params["bar"],
			"expected param 'bar' to have parameter source value")
		assert.Equal(t, "baz_default", params["baz"],
			"expected param 'baz' to have bundle default value")
	})
}

// This is intended to cover the matrix of cases around parameter value resolution.
// It exercises the matrix for all supported actions.
func Test_Paramapalooza(t *testing.T) {
	actions := []string{"install", "upgrade", "uninstall", "zombies"}
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

					bun := bundle.Bundle{
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
						Action:       action,
						Installation: "test",
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
						claim, err := claim.New("test", claim.ActionInstall, bun, nil)
						require.NoError(t, err)

						err = d.claims.SaveClaim(claim)
						require.NoError(t, err)
					}

					err := d.Execute(args)
					if tc.expectedErr != "" {
						require.EqualError(t, err, tc.expectedErr)
					} else {
						require.NoError(t, err)

						if action != "uninstall" {
							// Verify the updated param value on the generated claim
							updatedClaim, err := d.claims.ReadLastClaim("test")
							require.NoError(t, err)
							require.Equal(t, tc.expectedVal, updatedClaim.Parameters["my-param"])
						}
					}
				})
			}
		})
	}
}

func TestRuntime_ResolveParameterSources(t *testing.T) {
	r := NewTestRuntime(t)

	r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
	bun, err := r.ProcessBundle("bundle.json")
	require.NoError(t, err, "ProcessBundle failed")

	tc := r.claims.(*claims.TestClaimProvider)
	c := tc.CreateClaim("mybun", claim.ActionInstall, bun, nil)
	cr := tc.CreateResult(c, claim.StatusSucceeded)
	tc.CreateOutput(c, cr, "foo", []byte("abc123"))

	args := ActionArguments{
		Installation: "mybun",
	}
	got, err := r.resolveParameterSources(args)
	require.NoError(t, err, "resolveParameterSources failed")

	want := valuesource.Set{"foo": "abc123"}
	assert.Equal(t, want, got, "resolved incorrect parameter values")
}
