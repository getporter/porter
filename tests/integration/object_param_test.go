//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/stretchr/testify/require"
)

func TestInstall_objectParamWithAtPrefix(t *testing.T) {
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.AddTestBundleDir("testdata/bundles/bundle-with-object-params", false)

	installOpts := porter.NewInstallOptions()
	// Use @ prefix to load object parameter from file
	installOpts.Params = []string{
		"config=@./config.json",
	}

	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "Config name: test-config", "expected config.name value to be present in output")
}

func TestInstall_objectParamInline(t *testing.T) {
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.AddTestBundleDir("testdata/bundles/bundle-with-object-params", false)

	installOpts := porter.NewInstallOptions()
	// Pass inline JSON without @ prefix
	installOpts.Params = []string{
		`config={"name":"inline-config"}`,
	}

	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "Config name: inline-config", "expected inline config.name value to be present")
}

func TestInstall_objectParamWithParameterSet(t *testing.T) {
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.AddTestBundleDir("testdata/bundles/bundle-with-object-params", false)

	// Create a parameter set with @ prefix in value
	testParamSet := storage.NewParameterSet("", "myparam",
		secrets.SourceMap{
			Name: "config",
			Source: secrets.Source{
				Strategy: host.SourceValue,
				Hint:     "@./config.json",
			},
		},
	)

	p.TestParameters.InsertParameterSet(ctx, testParamSet)

	installOpts := porter.NewInstallOptions()
	installOpts.ParameterSets = []string{"myparam"}

	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "Config name: test-config", "expected config from parameter set to be used")
}

func TestInstall_objectParamFileNotFound(t *testing.T) {
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.AddTestBundleDir("testdata/bundles/bundle-with-object-params", false)

	installOpts := porter.NewInstallOptions()
	// Reference non-existent file
	installOpts.Params = []string{
		"config=@./nonexistent.json",
	}

	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.Error(t, err, "expected error when object parameter file not found")
	require.Contains(t, err.Error(), "unable to read file for object parameter config", "expected helpful error message")
}

func TestInstall_objectParamInvalidJSON(t *testing.T) {
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.AddTestBundleDir("testdata/bundles/bundle-with-object-params", false)

	// Create a file with invalid JSON in the test directory
	err := p.FileSystem.WriteFile("invalid.json", []byte("{this is not valid json}"), 0644)
	require.NoError(t, err)

	installOpts := porter.NewInstallOptions()
	installOpts.Params = []string{
		"config=@./invalid.json",
	}

	err = installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.Error(t, err, "expected error when object parameter file contains invalid JSON")
	require.Contains(t, err.Error(), "does not contain valid JSON", "expected helpful error message about invalid JSON")
}
