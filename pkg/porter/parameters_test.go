package porter

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/test"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisplayValuesSort(t *testing.T) {
	v := DisplayValues{
		{Name: "b"},
		{Name: "c"},
		{Name: "a"},
	}

	sort.Sort(v)

	assert.Equal(t, "a", v[0].Name)
	assert.Equal(t, "b", v[1].Name)
	assert.Equal(t, "c", v[2].Name)
}

func TestGenerateParameterSet(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

	opts := ParameterOptions{
		Silent: true,
	}
	opts.Namespace = "dev"
	opts.Name = "kool-params"
	opts.Labels = []string{"env=dev"}
	opts.CNABFile = "/bundle.json"
	ctx := context.Background()

	err := opts.Validate(ctx, nil, p.Porter)
	require.NoError(t, err, "Validate failed")

	err = p.GenerateParameters(ctx, opts)
	require.NoError(t, err, "no error should have existed")
	creds, err := p.Parameters.GetParameterSet(ctx, opts.Namespace, "kool-params")
	require.NoError(t, err, "expected parameter to have been generated")
	assert.Equal(t, map[string]string{"env": "dev"}, creds.Labels)
}

func TestPorter_ListParameters(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()
	ctx := context.Background()
	p.TestParameters.InsertParameterSet(ctx, storage.NewParameterSet("", "shared-mysql"))
	p.TestParameters.InsertParameterSet(ctx, storage.NewParameterSet("dev", "carolyn-wordpress"))
	p.TestParameters.InsertParameterSet(ctx, storage.NewParameterSet("dev", "vaughn-wordpress"))
	p.TestParameters.InsertParameterSet(ctx, storage.NewParameterSet("test", "staging-wordpress"))
	p.TestParameters.InsertParameterSet(ctx, storage.NewParameterSet("test", "iat-wordpress"))
	p.TestParameters.InsertParameterSet(ctx, storage.NewParameterSet("test", "shared-mysql"))

	t.Run("all-namespaces", func(t *testing.T) {
		opts := ListOptions{AllNamespaces: true}
		results, err := p.ListParameters(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 6)
	})

	t.Run("local namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: "dev"}
		results, err := p.ListParameters(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		opts = ListOptions{Namespace: "test"}
		results, err = p.ListParameters(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("global namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: ""}
		results, err := p.ListParameters(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func Test_loadParameters_paramNotDefined(t *testing.T) {
	t.Parallel()

	r := NewTestPorter(t)
	defer r.Close()

	b := cnab.NewBundle(bundle.Bundle{
		Parameters: map[string]bundle.Parameter{},
	})

	overrides := map[string]string{
		"foo": "bar",
	}

	i := storage.Installation{}
	_, err := r.resolveParameters(context.Background(), i, b, "action", overrides)
	require.EqualError(t, err, "parameter foo not defined in bundle")
}

func Test_loadParameters_definitionNotDefined(t *testing.T) {
	t.Parallel()

	r := NewTestPorter(t)
	defer r.Close()

	b := cnab.NewBundle(bundle.Bundle{
		Parameters: map[string]bundle.Parameter{
			"foo": {
				Definition: "foo",
			},
		},
	})

	overrides := map[string]string{
		"foo": "bar",
	}

	i := storage.Installation{}
	_, err := r.resolveParameters(context.Background(), i, b, "action", overrides)
	require.EqualError(t, err, "definition foo not defined in bundle")
}

func Test_loadParameters_applyTo(t *testing.T) {
	t.Parallel()

	r := NewTestPorter(t)
	defer r.Close()

	// Here we set default values, but expect nil/empty
	// values for parameters that do not apply to a given action
	b := cnab.NewBundle(bundle.Bundle{
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
	})

	overrides := map[string]string{
		"foo":  "FOO",
		"bar":  "456",
		"true": "false",
	}

	i := storage.Installation{}
	params, err := r.resolveParameters(context.Background(), i, b, "action", overrides)
	require.NoError(t, err)

	require.Equal(t, "FOO", params["foo"], "expected param 'foo' to be updated")
	require.EqualValues(t, 456, params["bar"], "expected param 'bar' to be updated")
	require.Equal(t, nil, params["true"], "expected param 'true' to be nil as it does not apply")
}

func Test_loadParameters_applyToBundleDefaults(t *testing.T) {
	t.Parallel()

	r := NewTestPorter(t)
	defer r.Close()

	b := cnab.NewBundle(bundle.Bundle{
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
	})

	i := storage.Installation{}
	params, err := r.resolveParameters(context.Background(), i, b, "action", nil)
	require.NoError(t, err)

	require.Equal(t, nil, params["foo"], "expected param 'foo' to be nil, regardless of the bundle default, as it does not apply")
}

func Test_loadParameters_requiredButDoesNotApply(t *testing.T) {
	t.Parallel()

	r := NewTestPorter(t)
	defer r.Close()

	b := cnab.NewBundle(bundle.Bundle{
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
	})

	i := storage.Installation{}
	params, err := r.resolveParameters(context.Background(), i, b, "action", nil)
	require.NoError(t, err)

	require.Equal(t, nil, params["foo"], "expected param 'foo' to be nil, regardless of claim value, as it does not apply")
}

func Test_loadParameters_fileParameter(t *testing.T) {
	t.Parallel()

	r := NewTestPorter(t)
	defer r.Close()

	r.TestConfig.TestContext.AddTestFile("testdata/file-param", "/path/to/file")

	b := cnab.NewBundle(bundle.Bundle{
		RequiredExtensions: []string{
			cnab.FileParameterExtensionKey,
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
	})

	overrides := map[string]string{
		"foo": "/path/to/file",
	}

	i := storage.Installation{}
	params, err := r.resolveParameters(context.Background(), i, b, "action", overrides)
	require.NoError(t, err)

	require.Equal(t, "SGVsbG8gV29ybGQh", params["foo"], "expected param 'foo' to be the base64-encoded file contents")
}

func Test_loadParameters_ParameterSourcePrecedence(t *testing.T) {
	t.Parallel()

	t.Run("nothing present, use default", func(t *testing.T) {
		t.Parallel()

		r := NewTestPorter(t)
		defer r.Close()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := cnab.LoadBundle(r.Context, "bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		i := storage.Installation{Name: "mybun"}
		params, err := r.resolveParameters(context.Background(), i, b, cnab.ActionUpgrade, nil)
		require.NoError(t, err)
		assert.Equal(t, "foo_default", params["foo"],
			"expected param 'foo' to have default value")
	})

	t.Run("only override present", func(t *testing.T) {
		t.Parallel()

		r := NewTestPorter(t)
		defer r.Close()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := cnab.LoadBundle(r.Context, "bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		overrides := map[string]string{
			"foo": "foo_override",
		}

		i := storage.Installation{Name: "mybun"}
		params, err := r.resolveParameters(context.Background(), i, b, cnab.ActionUpgrade, overrides)
		require.NoError(t, err)
		assert.Equal(t, "foo_override", params["foo"],
			"expected param 'foo' to have override value")
	})

	t.Run("only parameter source present", func(t *testing.T) {
		t.Parallel()

		r := NewTestPorter(t)
		defer r.Close()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := cnab.LoadBundle(r.Context, "bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		i := r.TestInstallations.CreateInstallation(storage.NewInstallation("", "mybun"))
		c := r.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall), func(r *storage.Run) { r.Bundle = b.Bundle })
		cr := r.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
		r.TestInstallations.CreateOutput(cr.NewOutput("foo", []byte("foo_source")))

		params, err := r.resolveParameters(context.Background(), i, b, cnab.ActionUpgrade, nil)
		require.NoError(t, err)
		assert.Equal(t, "foo_source", params["foo"],
			"expected param 'foo' to have parameter source value")
	})

	t.Run("override > parameter source", func(t *testing.T) {
		t.Parallel()

		r := NewTestPorter(t)
		defer r.Close()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := cnab.LoadBundle(r.Context, "bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		overrides := map[string]string{
			"foo": "foo_override",
		}

		i := r.TestInstallations.CreateInstallation(storage.NewInstallation("", "mybun"))
		c := r.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall))
		cr := r.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
		r.TestInstallations.CreateOutput(cr.NewOutput("foo", []byte("foo_source")))

		params, err := r.resolveParameters(context.Background(), i, b, cnab.ActionUpgrade, overrides)
		require.NoError(t, err)
		assert.Equal(t, "foo_override", params["foo"],
			"expected param 'foo' to have parameter override value")
	})

	t.Run("dependency output without type", func(t *testing.T) {
		t.Parallel()

		r := NewTestPorter(t)
		defer r.Close()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := cnab.LoadBundle(r.Context, "bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		i := r.TestInstallations.CreateInstallation(storage.NewInstallation("", "mybun-mysql"))
		c := r.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall), func(r *storage.Run) { r.Bundle = b.Bundle })
		cr := r.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
		r.TestInstallations.CreateOutput(cr.NewOutput("connstr", []byte("connstr value")))

		params, err := r.resolveParameters(context.Background(), i, b, cnab.ActionUpgrade, nil)
		require.NoError(t, err)
		assert.Equal(t, "connstr value", params["connstr"],
			"expected param 'connstr' to have parameter value from the untyped dependency output")
	})

	t.Run("merge parameter values", func(t *testing.T) {
		t.Parallel()

		r := NewTestPorter(t)
		defer r.Close()

		r.TestParameters.AddTestParameters("testdata/paramset.json")
		r.TestParameters.AddSecret("foo_secret", "foo_set")

		r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
		b, err := cnab.LoadBundle(r.Context, "bundle.json")
		require.NoError(t, err, "ProcessBundle failed")

		// foo is set by a user
		// bar is set by a parameter source
		// baz is set by the bundle default
		i := r.TestInstallations.CreateInstallation(storage.NewInstallation("", "mybun"))
		c := r.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall))
		cr := r.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
		r.TestInstallations.CreateOutput(cr.NewOutput("foo", []byte("foo_source")))
		r.TestInstallations.CreateOutput(cr.NewOutput("bar", []byte("bar_source")))
		r.TestInstallations.CreateOutput(cr.NewOutput("baz", []byte("baz_source")))

		overrides := map[string]string{"foo": "foo_override"}
		params, err := r.resolveParameters(context.Background(), i, b, cnab.ActionUpgrade, overrides)
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
			true, false, true, true, nil, "parameter \"my-param\" is required",
		},
		{"required, not provided, default exists, does not apply to action",
			true, false, true, false, nil, "",
		},
		{"required, not provided, default does not exist, applies to action",
			true, false, false, true, nil, "parameter \"my-param\" is required",
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
					t.Parallel()
					tc := tc

					r := NewTestPorter(t)
					defer r.Close()

					bun := cnab.NewBundle(bundle.Bundle{
						Name:          "mybuns",
						Version:       "1.0.0",
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
					})

					if tc.DefaultExists {
						bun.Definitions["my-param"].Default = "my-param-default"
					}

					if !tc.AppliesToAction {
						param := bun.Parameters["my-param"]
						param.ApplyTo = []string{"non-applicable-action"}
						bun.Parameters["my-param"] = param
					}

					i := storage.Installation{Name: "test"}
					overrides := map[string]string{}
					// If param is provided (via --param/--param-file)
					// it will be attached to args
					if tc.Provided {
						overrides["my-param"] = "my-param-value"
					}

					resolvedParams, err := r.resolveParameters(context.Background(), i, bun, action, overrides)
					if tc.ExpectedErr != "" {
						require.EqualError(t, err, tc.ExpectedErr)
					} else {
						require.NoError(t, err)
						assert.Equal(t, tc.ExpectedVal, resolvedParams["my-param"])
					}
				})
			}
		})
	}
}

func TestRuntime_ResolveParameterSources(t *testing.T) {
	t.Parallel()

	r := NewTestPorter(t)
	defer r.Close()

	r.TestConfig.TestContext.AddTestFile("testdata/bundle-with-param-sources.json", "bundle.json")
	bun, err := cnab.LoadBundle(r.Context, "bundle.json")
	require.NoError(t, err, "ProcessBundle failed")

	i := r.TestInstallations.CreateInstallation(storage.NewInstallation("", "mybun-mysql"))
	c := r.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall), func(r *storage.Run) { r.Bundle = bun.Bundle })
	cr := r.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
	r.TestInstallations.CreateOutput(cr.NewOutput("connstr", []byte("connstr value")))

	i = r.TestInstallations.CreateInstallation(storage.NewInstallation("", "mybun"))
	c = r.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall), func(r *storage.Run) { r.Bundle = bun.Bundle })
	cr = r.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
	r.TestInstallations.CreateOutput(cr.NewOutput("bar", []byte("bar value")))

	got, err := r.resolveParameterSources(context.Background(), bun, i)
	require.NoError(t, err, "resolveParameterSources failed")

	want := secrets.Set{
		"bar":     "bar value",
		"connstr": "connstr value",
	}
	assert.Equal(t, want, got, "resolved incorrect parameter values")
}

func TestShowParameters_NotFound(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := ParameterShowOptions{
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatPlaintext,
		},
		Name: "non-existent-param",
	}

	err := p.ShowParameter(context.Background(), opts)
	assert.ErrorIs(t, err, storage.ErrNotFound{})
}

func TestShowParameters_Found(t *testing.T) {
	type ParameterShowTest struct {
		name               string
		format             printer.Format
		expectedOutputFile string
	}

	testcases := []ParameterShowTest{
		{
			name:               "json",
			format:             printer.FormatJson,
			expectedOutputFile: "testdata/parameters/mypset.json",
		},
		{
			name:               "yaml",
			format:             printer.FormatYaml,
			expectedOutputFile: "testdata/parameters/mypset.yaml",
		},
		{
			name:               "plaintext",
			format:             printer.FormatPlaintext,
			expectedOutputFile: "testdata/parameters/mypset.txt",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			opts := ParameterShowOptions{
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
				Name: "mypset",
			}

			p.TestParameters.AddTestParameters("testdata/paramset.json")

			err := p.ShowParameter(context.Background(), opts)
			require.NoError(t, err, "an error should not have occurred")
			gotOutput := p.TestConfig.TestContext.GetOutput()
			test.CompareGoldenFile(t, tc.expectedOutputFile, gotOutput)
		})
	}
}

func TestParametersCreateOptions_Validate(t *testing.T) {
	testcases := []struct {
		name       string
		args       []string
		outputType string
		wantErr    string
	}{
		{
			name:       "no fileName defined",
			args:       []string{},
			outputType: "",
			wantErr:    "",
		},
		{
			name:       "two positional arguments",
			args:       []string{"parameter-set1", "parameter-set2"},
			outputType: "",
			wantErr:    "only one positional argument may be specified",
		},
		{
			name:       "no file format defined from file extension or output flag",
			args:       []string{"parameter-set"},
			outputType: "",
			wantErr:    "could not detect the file format from the file extension (.txt). Specify the format with --output.",
		},
		{
			name:       "different file format",
			args:       []string{"parameter-set.json"},
			outputType: "yaml",
			wantErr:    "",
		},
		{
			name:       "format from output flag",
			args:       []string{"parameters"},
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "format from file extension",
			args:       []string{"parameter-set.yml"},
			outputType: "",
			wantErr:    "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := ParameterCreateOptions{OutputType: tc.outputType}
			err := opts.Validate(tc.args)
			if tc.wantErr == "" {
				require.NoError(t, err, "no error should have existed")
				return
			}
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestParametersCreate(t *testing.T) {
	testcases := []struct {
		name       string
		fileName   string
		outputType string
		wantErr    string
	}{
		{
			name:       "valid input: no input defined, will output yaml format to stdout",
			fileName:   "",
			outputType: "",
			wantErr:    "",
		},
		{
			name:       "valid input: output to stdout with format json",
			fileName:   "",
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "valid input: file format from fileName",
			fileName:   "fileName.json",
			outputType: "",
			wantErr:    "",
		},
		{
			name:       "valid input: file format from outputType",
			fileName:   "fileName",
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "valid input: different file format from fileName and outputType",
			fileName:   "fileName.yaml",
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "valid input: same file format in fileName and outputType",
			fileName:   "fileName.json",
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "invalid input: invalid file format from fileName",
			fileName:   "fileName.txt",
			outputType: "",
			wantErr:    fmt.Sprintf("unsupported format %s. Supported formats are: yaml and json.", "txt"),
		},
		{
			name:       "invalid input: invalid file format from outputType",
			fileName:   "fileName",
			outputType: "txt",
			wantErr:    fmt.Sprintf("unsupported format %s. Supported formats are: yaml and json.", "txt"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			opts := ParameterCreateOptions{FileName: tc.fileName, OutputType: tc.outputType}
			err := p.CreateParameter(opts)
			if tc.wantErr == "" {
				require.NoError(t, err, "no error should have existed")
				return
			}
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
