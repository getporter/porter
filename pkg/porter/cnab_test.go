package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedOptions_defaultBundleFiles(t *testing.T) {
	cxt := portercontext.NewTestContext(t)

	_, err := cxt.FileSystem.Create("porter.yaml")
	require.NoError(t, err)

	opts := sharedOptions{}
	err = opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, "porter.yaml", opts.File)
	assert.Equal(t, ".cnab/bundle.json", opts.CNABFile)
}

func TestSharedOptions_defaultBundleFiles_AltManifest(t *testing.T) {
	cxt := portercontext.NewTestContext(t)

	opts := sharedOptions{
		bundleFileOptions: bundleFileOptions{
			File: "mybun/porter.yaml",
		},
	}
	err := opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, ".cnab/bundle.json", opts.CNABFile)
}

func TestSharedOptions_defaultBundleFiles_CNABFile(t *testing.T) {
	cxt := portercontext.NewTestContext(t)

	// Add existing porter manifest; ensure it isn't processed when cnab-file is spec'd
	_, err := cxt.FileSystem.Create("porter.yaml")
	_, err = cxt.FileSystem.Create("mycnabfile.json")
	require.NoError(t, err)

	opts := sharedOptions{}
	opts.CNABFile = "mycnabfile.json"
	err = opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, "", opts.File)
	assert.Equal(t, "mycnabfile.json", opts.CNABFile)
}

func TestSharedOptions_validateBundleJson(t *testing.T) {
	cxt := portercontext.NewTestContext(t)

	cxt.FileSystem.Create("mybun1/bundle.json")
	cxt.FileSystem.Create("bundle1.json")

	testcases := []struct {
		name           string
		cnabFile       string
		wantBundleJson string
		wantError      string
	}{
		{name: "absolute file exists", cnabFile: "/mybun1/bundle.json", wantBundleJson: "/mybun1/bundle.json", wantError: ""},
		{name: "relative file exists", cnabFile: "bundle1.json", wantBundleJson: "/bundle1.json", wantError: ""},
		{name: "absolute file does not exist", cnabFile: "mybun2/bundle.json", wantError: "unable to access --cnab-file mybun2/bundle.json"},
		{name: "relative file does not", cnabFile: "bundle2.json", wantError: "unable to access --cnab-file bundle2.json"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := sharedOptions{
				bundleFileOptions: bundleFileOptions{
					CNABFile: tc.cnabFile,
				},
			}

			err := opts.validateCNABFile(cxt.Context)

			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, opts.CNABFile, tc.wantBundleJson)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError)
			}
		})
	}
}

func TestSharedOptions_defaultDriver(t *testing.T) {
	opts := sharedOptions{}

	opts.defaultDriver()

	assert.Equal(t, DefaultDriver, opts.Driver)
}

func TestSharedOptions_ParseParamSets(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestParameters.AddSecret("foo_secret", "foo_value")
	p.TestParameters.AddSecret("PARAM2_SECRET", "VALUE2")
	p.TestParameters.AddTestParameters("testdata/paramset2.json")

	opts := sharedOptions{
		ParameterSets: []string{
			"porter-hello",
		},
	}

	err := opts.Validate(context.Background(), []string{}, p.Porter)
	assert.NoError(t, err)

	err = opts.parseParamSets(context.Background(), p.Porter, cnab.ExtendedBundle{})
	assert.NoError(t, err)

	wantParams := map[string]string{
		"my-second-param": "VALUE2",
	}
	assert.Equal(t, wantParams, opts.parsedParamSets, "resolved unexpected parameter values")
}

func TestSharedOptions_ParseParamSets_Failed(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/porter-with-file-param.yaml", config.Name)
	p.TestConfig.TestContext.AddTestFile("testdata/paramset-with-file-param.json", "/paramset.json")

	m, err := manifest.LoadManifestFrom(context.Background(), p.Config, config.Name)
	require.NoError(t, err)
	bun, err := configadapter.ConvertToTestBundle(p.Context, m)
	require.NoError(t, err)

	opts := sharedOptions{
		ParameterSets: []string{
			"/paramset.json",
		},
		bundleFileOptions: bundleFileOptions{
			File: "porter.yaml",
		},
	}

	err = opts.Validate(context.Background(), []string{}, p.Porter)
	assert.NoError(t, err)

	err = opts.parseParamSets(context.Background(), p.Porter, bun)
	assert.Error(t, err)

}

func TestSharedOptions_LoadParameters(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", config.Name)
	m, err := manifest.LoadManifestFrom(context.Background(), p.Config, config.Name)
	require.NoError(t, err)
	bun, err := configadapter.ConvertToTestBundle(p.Context, m)
	require.NoError(t, err)

	opts := sharedOptions{}
	opts.Params = []string{"my-first-param=1", "my-second-param=2"}

	err = opts.LoadParameters(context.Background(), p.Porter, bun)
	require.NoError(t, err)

	assert.Len(t, opts.Params, 2)
}

func TestSharedOptions_CombineParameters(t *testing.T) {
	c := portercontext.NewTestContext(t)
	c.Debug = false

	t.Run("no override present, no parameter set present", func(t *testing.T) {
		opts := sharedOptions{}

		params := opts.combineParameters(c.Context)
		require.Equal(t, map[string]string{}, params,
			"expected combined params to be empty")
	})

	t.Run("override present, no parameter set present", func(t *testing.T) {
		opts := sharedOptions{
			parsedParams: map[string]string{
				"foo": "foo_cli_override",
			},
		}

		params := opts.combineParameters(c.Context)
		require.Equal(t, "foo_cli_override", params["foo"],
			"expected param 'foo' to have override value")
	})

	t.Run("no override present, parameter set present", func(t *testing.T) {
		opts := sharedOptions{
			parsedParamSets: map[string]string{
				"foo": "foo_via_paramset",
			},
		}

		params := opts.combineParameters(c.Context)
		require.Equal(t, "foo_via_paramset", params["foo"],
			"expected param 'foo' to have parameter set value")
	})

	t.Run("override present, parameter set present", func(t *testing.T) {
		opts := sharedOptions{
			parsedParams: map[string]string{
				"foo": "foo_cli_override",
			},
			parsedParamSets: map[string]string{
				"foo": "foo_via_paramset",
			},
		}

		params := opts.combineParameters(c.Context)
		require.Equal(t, "foo_cli_override", params["foo"],
			"expected param 'foo' to have override value, which has precedence over the parameter set value")
	})

	t.Run("debug on", func(t *testing.T) {
		var opts sharedOptions
		debugContext := portercontext.NewTestContext(t)
		debugContext.Debug = true
		params := opts.combineParameters(debugContext.Context)
		require.Equal(t, "true", params["porter-debug"], "porter-debug should be set to true when p.Debug is true")
	})
}

func Test_bundleFileOptions(t *testing.T) {
	testcases := []struct {
		name         string
		opts         bundleFileOptions
		setup        func(*portercontext.Context, bundleFileOptions) error
		wantFile     string
		wantCNABFile string
		wantError    string
	}{
		{
			name:         "no opts",
			opts:         bundleFileOptions{},
			setup:        func(ctx *portercontext.Context, opts bundleFileOptions) error { return nil },
			wantFile:     config.Name,
			wantCNABFile: build.LOCAL_BUNDLE,
			wantError:    "",
		}, {
			name: "reference set",
			opts: bundleFileOptions{
				ReferenceSet: true,
			},
			setup:        func(ctx *portercontext.Context, opts bundleFileOptions) error { return nil },
			wantFile:     "",
			wantCNABFile: "",
			wantError:    "",
		}, {
			name: "invalid dir",
			opts: bundleFileOptions{
				Dir: "path/to/bundle",
			},
			setup:        func(ctx *portercontext.Context, opts bundleFileOptions) error { return nil },
			wantFile:     "",
			wantCNABFile: "",
			wantError:    `"path/to/bundle" is not a valid directory: open /path/to/bundle: file does not exist`,
		}, {
			name: "invalid file",
			opts: bundleFileOptions{
				File: "alternate/porter.yaml",
			},
			setup:        func(ctx *portercontext.Context, opts bundleFileOptions) error { return nil },
			wantFile:     "",
			wantCNABFile: "",
			wantError:    "unable to access --file alternate/porter.yaml: open /alternate/porter.yaml: file does not exist",
		}, {
			name: "valid dir",
			opts: bundleFileOptions{
				Dir: "path/to/bundle",
			},
			setup: func(ctx *portercontext.Context, opts bundleFileOptions) error {
				return ctx.FileSystem.MkdirAll(opts.Dir, pkg.FileModeDirectory)
			},
			wantFile:     config.Name,
			wantCNABFile: "/path/to/bundle/.cnab/bundle.json",
			wantError:    "",
		}, {
			name: "valid file",
			opts: bundleFileOptions{
				File: "alternate/porter.yaml",
			},
			setup: func(ctx *portercontext.Context, opts bundleFileOptions) error {
				return ctx.FileSystem.MkdirAll(opts.File, pkg.FileModeDirectory)
			},
			wantFile:     "/alternate/porter.yaml",
			wantCNABFile: build.LOCAL_BUNDLE,
			wantError:    "",
		}, {
			name: "valid dir and file",
			opts: bundleFileOptions{
				Dir:  "path/to/bundle",
				File: "alternate/porter.yaml",
			},
			setup: func(ctx *portercontext.Context, opts bundleFileOptions) error {
				err := ctx.FileSystem.MkdirAll(opts.File, pkg.FileModeDirectory)
				if err != nil {
					return err
				}
				return ctx.FileSystem.MkdirAll(opts.Dir, pkg.FileModeDirectory)
			},
			wantFile:     "/alternate/porter.yaml",
			wantCNABFile: "/path/to/bundle/.cnab/bundle.json",
			wantError:    "",
		}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cxt := portercontext.NewTestContext(t)

			// Create default local manifest
			_, err := cxt.FileSystem.Create(config.Name)
			require.NoError(t, err)

			err = tc.setup(cxt.Context, tc.opts)
			require.NoError(t, err)

			err = tc.opts.Validate(cxt.Context)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err)

				require.Equal(t, tc.wantFile, tc.opts.File)
				require.Equal(t, tc.wantCNABFile, tc.opts.CNABFile)

				// Working Dir assertions
				wd := cxt.FileSystem.Getwd()
				if tc.opts.Dir != "" && tc.wantError == "" {
					require.Equal(t, tc.opts.Dir, wd)
				} else {
					require.Equal(t, "/", wd)
				}
			}
		})
	}
}

func TestSharedOptions_populateInternalParameterSet(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", config.Name)
	m, err := manifest.LoadManifestFrom(context.Background(), p.Config, config.Name)
	require.NoError(t, err)
	bun, err := configadapter.ConvertToTestBundle(p.Context, m)
	require.NoError(t, err)

	sensitiveParamName := "my-second-param"
	sensitiveParamValue := "2"
	nonsensitiveParamName := "my-first-param"
	nonsensitiveParamValue := "1"
	opts := sharedOptions{}
	opts.Params = []string{nonsensitiveParamName + "=" + nonsensitiveParamValue, sensitiveParamName + "=" + sensitiveParamValue}

	err = opts.LoadParameters(p.Porter, bun)
	require.NoError(t, err)

	i := claims.NewInstallation("", bun.Name)

	err = opts.populateInternalParameterSet(p.Porter, bun, &i)
	require.NoError(t, err)

	require.Len(t, i.Parameters.Parameters, 2)

	// there should be no sensitive value on installation record
	for _, param := range i.Parameters.Parameters {
		if param.Name == sensitiveParamName {
			require.Equal(t, param.Source.Key, secrets.SourceSecret)
			require.NotEqual(t, param.Source.Value, sensitiveParamValue)
			continue
		}
		require.Equal(t, param.Source.Key, host.SourceValue)
		require.Equal(t, param.Source.Value, nonsensitiveParamValue)
	}

	// if no parameter override specified, installation record should be updated
	// as well
	opts.combinedParameters = nil
	opts.Params = make([]string, 0)
	err = opts.LoadParameters(p.Porter, bun)
	require.NoError(t, err)
	err = opts.populateInternalParameterSet(p.Porter, bun, &i)
	require.NoError(t, err)

	require.Len(t, i.Parameters.Parameters, 0)
}
