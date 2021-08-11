package cnabprovider

import (
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/cnab/extensions"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_loadParameters_paramNotDefined(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Teardown()

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
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Teardown()

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
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Teardown()

	// Here we set default values, but expect nil/empty
	// values for parameters that do not apply to a given action
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
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Teardown()

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
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Teardown()

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
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Teardown()

	r.TestConfig.TestContext.AddTestFile("testdata/file-param", "/path/to/file")

	b := bundle.Bundle{
		RequiredExtensions: []string{
			extensions.FileParameterExtensionKey,
		},
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
	t.Parallel()

	t.Run("nothing present, use default", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Teardown()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.TestSecrets.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := r.ProcessBundleFromFile("bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		args := ActionArguments{
			Installation: "mybun",
			Action:       cnab.ActionUpgrade,
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "foo_default", params["foo"],
			"expected param 'foo' to have default value")
	})

	t.Run("only override present", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Teardown()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.TestSecrets.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := r.ProcessBundleFromFile("bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		overrides := map[string]string{
			"foo": "foo_override",
		}

		args := ActionArguments{
			Installation: "mybun",
			Action:       cnab.ActionUpgrade,
			Params:       overrides,
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "foo_override", params["foo"],
			"expected param 'foo' to have override value")
	})

	t.Run("only parameter source present", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Teardown()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.TestSecrets.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := r.ProcessBundleFromFile("bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		i := r.TestClaims.CreateInstallation(claims.NewInstallation("", "mybun"))
		c := r.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall), func(r *claims.Run) { r.Bundle = b })
		cr := r.TestClaims.CreateResult(c.NewResult(cnab.StatusSucceeded))
		r.TestClaims.CreateOutput(cr.NewOutput("foo", []byte("foo_source")))

		args := ActionArguments{
			Installation: "mybun",
			Action:       cnab.ActionUpgrade,
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "foo_source", params["foo"],
			"expected param 'foo' to have parameter source value")
	})

	t.Run("override > parameter source", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Teardown()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.TestSecrets.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := r.ProcessBundleFromFile("bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		overrides := map[string]string{
			"foo": "foo_override",
		}

		i := r.TestClaims.CreateInstallation(claims.NewInstallation("", "mybun"))
		c := r.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall))
		cr := r.TestClaims.CreateResult(c.NewResult(cnab.StatusSucceeded))
		r.TestClaims.CreateOutput(cr.NewOutput("foo", []byte("foo_source")))

		args := ActionArguments{
			Installation: "mybun",
			Action:       cnab.ActionUpgrade,
			Params:       overrides,
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "foo_override", params["foo"],
			"expected param 'foo' to have parameter override value")
	})

	t.Run("dependency output without type", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Teardown()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.TestSecrets.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := r.ProcessBundleFromFile("bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		bun := bundle.Bundle{
			Name:    "foo-setup",
			Version: bundle.CNABSpecVersion,
			InvocationImages: []bundle.InvocationImage{
				{
					BaseImage: bundle.BaseImage{
						ImageType: "docker",
						Image:     "getporter/foo-setup:latest",
					},
				},
			},
			Outputs: map[string]bundle.Output{
				"connstr": {Definition: "connstr"}},
			Definitions: map[string]*definition.Schema{
				"connstr": {Type: "string"},
			},
		}

		i := r.TestClaims.CreateInstallation(claims.NewInstallation("", "mybun-mysql"))
		c := r.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall), func(r *claims.Run) { r.Bundle = bun })
		cr := r.TestClaims.CreateResult(c.NewResult(cnab.StatusSucceeded))
		r.TestClaims.CreateOutput(cr.NewOutput("connstr", []byte("connstr value")))

		args := ActionArguments{
			Installation: "mybun",
			Action:       cnab.ActionUpgrade,
		}
		params, err := r.loadParameters(b, args)
		require.NoError(t, err)
		assert.Equal(t, "connstr value", params["connstr"],
			"expected param 'connstr' to have parameter value from the untyped dependency output")
	})

	t.Run("merge parameter values", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Teardown()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.TestSecrets.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := r.ProcessBundleFromFile("bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		// foo is set by a user
		// bar is set by a parameter source
		// baz is set by the bundle default
		i := r.TestClaims.CreateInstallation(claims.NewInstallation("", "mybun"))
		c := r.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall))
		cr := r.TestClaims.CreateResult(c.NewResult(cnab.StatusSucceeded))
		r.TestClaims.CreateOutput(cr.NewOutput("foo", []byte("foo_source")))
		r.TestClaims.CreateOutput(cr.NewOutput("bar", []byte("bar_source")))
		r.TestClaims.CreateOutput(cr.NewOutput("baz", []byte("baz_source")))

		args := ActionArguments{
			Installation: "mybun",
			Action:       cnab.ActionUpgrade,
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
	testcases := []struct {
		Name            string
		Required        bool
		Provided        bool
		DefaultExists   bool
		AppliesToAction bool
		ExpectedVal     interface{}
		ExpectedErr     string
	}{
		// Are you ready to enter the Matrix?
		{Name: "required, provided, default exists, applies to action",
			Required: true, Provided: true, DefaultExists: true, AppliesToAction: true, ExpectedVal: "my-param-value",
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
		t.Run(tc.Name, func(t *testing.T) {
			actions := []string{"install", "upgrade", "uninstall", "zombies"}
			for _, action := range actions {
				t.Run(action, func(t *testing.T) {
					//t.Parallel()
					tc := tc

					r := NewTestRuntime(t)
					defer r.Teardown()

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
								Required:   tc.Required,
								Destination: &bundle.Location{
									EnvironmentVariable: "MY_PARAM",
								},
							},
						},
					}

					if tc.DefaultExists {
						bun.Definitions["my-param"].Default = "my-param-default"
					}

					if !tc.AppliesToAction {
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
					if tc.Provided {
						args.Params = map[string]string{
							"my-param": "my-param-value",
						}
					}

					// If action is install, no claim is expected to exist
					// so we write a bundle and pull it in via ActionArguments
					if action == "install" {
						bytes, err := json.Marshal(bun)
						require.NoError(t, err)

						err = r.FileSystem.WriteFile("bundle.json", bytes, 0644)
						require.NoError(t, err)

						args.BundlePath = "bundle.json"
					} else {
						// For all other actions, a claim is expected to exist
						// so we create one here and add the bundle to the claim
						i := claims.NewInstallation("", "test")
						require.NoError(t, r.claims.InsertInstallation(i), "error creating existing testing installation")
						c := i.NewRun(cnab.ActionInstall)
						c.Bundle = bun
						require.NoError(t, r.claims.InsertRun(c), "error adding existing test run")
						require.NoError(t, r.claims.InsertResult(c.NewResult(cnab.StatusSucceeded)), "error setting existing run status")
					}

					err := r.Execute(args)
					if tc.ExpectedErr != "" {
						require.EqualError(t, err, tc.ExpectedErr)
					} else {
						require.NoError(t, err)

						if action != "uninstall" {
							// Verify the updated param value on the generated claim
							updatedClaim, err := r.claims.GetLastRun("", "test")
							require.NoError(t, err)
							require.Equal(t, tc.ExpectedVal, updatedClaim.Parameters["my-param"])
						}
					}
				})
			}
		})
	}
}

func TestRuntime_ResolveParameterSources(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Teardown()

	r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
	bun, err := r.ProcessBundleFromFile("bundle.json")
	require.NoError(t, err, "ProcessBundle failed")

	fooBun := bundle.Bundle{
		Name:    "foo-setup",
		Version: bundle.CNABSpecVersion,
		InvocationImages: []bundle.InvocationImage{
			{
				BaseImage: bundle.BaseImage{
					ImageType: "docker",
					Image:     "getporter/foo-setup:latest",
				},
			},
		},
		Outputs: map[string]bundle.Output{
			"connstr": {Definition: "connstr"}},
		Definitions: map[string]*definition.Schema{
			"connstr": {Type: "string"},
		},
	}
	i := r.TestClaims.CreateInstallation(claims.NewInstallation("", "mybun-mysql"))
	c := r.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall), func(r *claims.Run) { r.Bundle = fooBun })
	cr := r.TestClaims.CreateResult(c.NewResult(cnab.StatusSucceeded))
	r.TestClaims.CreateOutput(cr.NewOutput("connstr", []byte("connstr value")))

	i = r.TestClaims.CreateInstallation(claims.NewInstallation("", "mybun"))
	c = r.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall), func(r *claims.Run) { r.Bundle = bun })
	cr = r.TestClaims.CreateResult(c.NewResult(cnab.StatusSucceeded))
	r.TestClaims.CreateOutput(cr.NewOutput("bar", []byte("bar value")))

	args := ActionArguments{
		Installation: "mybun",
	}
	got, err := r.resolveParameterSources(bun, args)
	require.NoError(t, err, "resolveParameterSources failed")

	want := secrets.Set{
		"bar":     "bar value",
		"connstr": "connstr value",
	}
	assert.Equal(t, want, got, "resolved incorrect parameter values")
}
