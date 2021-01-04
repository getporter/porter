package porter

import (
	"fmt"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedOptions_defaultBundleFiles(t *testing.T) {
	cxt := context.NewTestContext(t)

	_, err := cxt.FileSystem.Create("porter.yaml")
	require.NoError(t, err)

	opts := sharedOptions{}
	err = opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, "porter.yaml", opts.File)
	assert.Equal(t, filepath.FromSlash(".cnab/bundle.json"), opts.CNABFile)
}

func TestSharedOptions_defaultBundleFiles_AltManifest(t *testing.T) {
	cxt := context.NewTestContext(t)

	opts := sharedOptions{
		bundleFileOptions: bundleFileOptions{
			File: "mybun/porter.yaml",
		},
	}
	err := opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, filepath.FromSlash("mybun/.cnab/bundle.json"), opts.CNABFile)
}

func TestSharedOptions_defaultBundleFiles_CNABFile(t *testing.T) {
	cxt := context.NewTestContext(t)

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
	cxt := context.NewTestContext(t)

	cxt.FileSystem.Create("mybun1/bundle.json")
	cxt.FileSystem.Create("bundle1.json")

	goodAbs := cxt.FileSystem.Abs("/mybun1/bundle.json")
	missingAbs := cxt.FileSystem.Abs("/mybun2/bundle.json")
	testcases := []struct {
		name           string
		cnabFile       string
		wantBundleJson string
		wantError      string
	}{
		{name: "absolute file exists", cnabFile: goodAbs, wantBundleJson: goodAbs, wantError: ""},
		{name: "relative file exists", cnabFile: "bundle1.json", wantBundleJson: cxt.FileSystem.Abs("/bundle1.json"), wantError: ""},
		{name: "absolute file does not exist", cnabFile: missingAbs, wantError: fmt.Sprintf("unable to access --cnab-file %s", missingAbs)},
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

func TestSharedOptions_ParseParamSets_viaPathOrName(t *testing.T) {
	p := NewTestPorter(t)

	paramSet1File := filepath.Join(p.Getwd(), "paramset.json")
	p.TestParameters.TestSecrets.AddSecret("foo_secret", "foo_value")
	p.TestParameters.TestSecrets.AddSecret("PARAM2_SECRET", "VALUE2")
	p.TestConfig.TestContext.AddTestFile("testdata/paramset.json", paramSet1File)
	p.TestParameters.AddTestParameters("testdata/paramset2.json")

	opts := sharedOptions{
		ParameterSets: []string{
			"porter-hello",
			paramSet1File,
		},
	}

	err := opts.Validate([]string{}, p.Porter)
	assert.NoError(t, err)

	err = opts.parseParamSets(p.Porter)
	assert.NoError(t, err)

	wantParams := map[string]string{
		"PARAM2": "VALUE2",
		"foo":    "foo_value",
	}
	assert.Equal(t, wantParams, opts.parsedParamSets, "resolved unexpected parameter values")
}

func TestSharedOptions_ParseParamSets_FileType(t *testing.T) {
	p := NewTestPorter(t)

	paramFile := filepath.Join(p.Getwd(), "paramset.json")
	p.TestConfig.TestContext.AddTestFile("testdata/porter-with-file-param.yaml", "porter.yaml")
	p.TestConfig.TestContext.AddTestFile("testdata/paramset-with-file-param.json", paramFile)

	opts := sharedOptions{
		ParameterSets: []string{
			paramFile,
		},
		bundleFileOptions: bundleFileOptions{
			File: "porter.yaml",
		},
	}

	err := opts.Validate([]string{}, p.Porter)
	assert.NoError(t, err)

	err = opts.parseParamSets(p.Porter)
	assert.NoError(t, err)

	wantParams := map[string]string{
		"my-file-param": "/local/path/to/my-file-param",
	}
	assert.Equal(t, wantParams, opts.parsedParamSets, "resolved unexpected parameter values")
}

func TestSharedOptions_LoadParameters(t *testing.T) {
	p := NewTestPorter(t)
	opts := sharedOptions{}
	opts.Params = []string{"A=1", "B=2"}

	err := opts.LoadParameters(p.Porter)
	require.NoError(t, err)

	assert.Len(t, opts.Params, 2)
}

func TestSharedOptions_CombineParameters(t *testing.T) {
	c := context.NewTestContext(t)
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
		debugContext := context.NewTestContext(t)
		debugContext.Debug = true
		params := opts.combineParameters(debugContext.Context)
		require.Equal(t, "true", params["porter-debug"], "porter-debug should be set to true when p.Debug is true")
	})
}
