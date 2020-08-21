package porter

import (
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedOptions_defaultBundleFiles(t *testing.T) {
	cxt := context.NewTestContext(t)

	pwd, _ := os.Getwd()
	_, err := cxt.FileSystem.Create(filepath.Join(pwd, "porter.yaml"))
	require.NoError(t, err)

	opts := sharedOptions{}
	err = opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, "porter.yaml", opts.File)
	assert.Equal(t, build.LOCAL_BUNDLE, opts.CNABFile)
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

	assert.Equal(t, filepath.Join("mybun", build.LOCAL_BUNDLE), opts.CNABFile)
}

func TestSharedOptions_defaultBundleFiles_CNABFile(t *testing.T) {
	cxt := context.NewTestContext(t)

	pwd, _ := os.Getwd()
	// Add existing porter manifest; ensure it isn't processed when cnab-file is spec'd
	_, err := cxt.FileSystem.Create(filepath.Join(pwd, "porter.yaml"))
	_, err = cxt.FileSystem.Create(filepath.Join(pwd, "mycnabfile.json"))
	require.NoError(t, err)

	opts := sharedOptions{}
	opts.CNABFile = "mycnabfile.json"
	err = opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, "", opts.File)
	assert.Equal(t, "mycnabfile.json", opts.CNABFile)
}

func TestSharedOptions_validateBundleJson(t *testing.T) {
	pwd, _ := os.Getwd()

	cxt := context.NewTestContext(t)

	cxt.FileSystem.Create("/mybun1/bundle.json")
	cxt.FileSystem.Create(filepath.Join(pwd, "bundle1.json"))

	testcases := []struct {
		name           string
		cnabFile       string
		wantBundleJson string
		wantError      string
	}{
		{name: "absolute file exists", cnabFile: "/mybun1/bundle.json", wantBundleJson: "/mybun1/bundle.json", wantError: ""},
		{name: "relative file exists", cnabFile: "bundle1.json", wantBundleJson: filepath.Join(pwd, "bundle1.json"), wantError: ""},
		{name: "absolute file does not exist", cnabFile: "/mybun2/bundle.json", wantError: "unable to access --cnab-file /mybun2/bundle.json"},
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

func TestParseParamSets_viaPathOrName(t *testing.T) {
	p := NewTestPorter(t)

	p.TestParameters.TestSecrets.AddSecret("foo_secret", "foo_value")
	p.TestParameters.TestSecrets.AddSecret("PARAM2_SECRET", "VALUE2")
	p.TestConfig.TestContext.AddTestFile("testdata/paramset.json", "/paramset.json")
	p.TestParameters.AddTestParameters("testdata/paramset2.json")

	opts := sharedOptions{
		ParameterSets: []string{
			"HELLO_CUSTOM",
			"/paramset.json",
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

func TestParseParamSets_FileType(t *testing.T) {
	p := NewTestPorter(t)

	p.TestConfig.TestContext.AddTestFile("testdata/porter-with-file-param.yaml", "porter.yaml")
	p.TestConfig.TestContext.AddTestFile("testdata/paramset-with-file-param.json", "/paramset.json")

	opts := sharedOptions{
		ParameterSets: []string{
			"/paramset.json",
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

func TestCombineParameters(t *testing.T) {
	t.Run("no override present, no parameter set present", func(t *testing.T) {
		opts := sharedOptions{}

		params := opts.combineParameters()
		require.Equal(t, map[string]string{}, params,
			"expected combined params to be empty")
	})

	t.Run("override present, no parameter set present", func(t *testing.T) {
		opts := sharedOptions{
			parsedParams: map[string]string{
				"foo": "foo_cli_override",
			},
		}

		params := opts.combineParameters()
		require.Equal(t, "foo_cli_override", params["foo"],
			"expected param 'foo' to have override value")
	})

	t.Run("no override present, parameter set present", func(t *testing.T) {
		opts := sharedOptions{
			parsedParamSets: map[string]string{
				"foo": "foo_via_paramset",
			},
		}

		params := opts.combineParameters()
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

		params := opts.combineParameters()
		require.Equal(t, "foo_cli_override", params["foo"],
			"expected param 'foo' to have override value, which has precedence over the parameter set value")
	})
}
