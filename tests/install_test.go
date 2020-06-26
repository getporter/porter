// +build integration

package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
)

func TestInstall_relativePathPorterHome(t *testing.T) {
	p := porter.NewTestPorter(t)

	// Crux for this test: set Porter's home dir to a relative path
	homeDir, err := p.Config.GetHomeDir()
	require.NoError(t, err)
	curDir, err := os.Getwd()
	require.NoError(t, err)
	relDir, err := filepath.Rel(curDir, homeDir)
	require.NoError(t, err)
	os.Setenv(config.EnvHOME, relDir)

	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	// Bring in a porter manifest that has an install action defined
	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/bundle-with-custom-action.yaml"), "porter.yaml")
	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/helpers.sh"), "helpers.sh")

	installOpts := porter.InstallOptions{}
	err = installOpts.Validate([]string{}, p.CNAB, p.Context)
	require.NoError(t, err)

	// Install the bundle, assert no error occurs due to Porter home as relative path
	err = p.InstallBundle(installOpts)
	require.NoError(t, err)
}

func TestInstall_fileParam(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/bundle-with-file-params.yaml"), "porter.yaml")
	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/helpers.sh"), "helpers.sh")
	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/myfile"), "./myfile")

	installOpts := porter.InstallOptions{}
	installOpts.Params = []string{"myfile=./myfile"}

	err := installOpts.Validate([]string{}, p.CNAB, p.Context)
	require.NoError(t, err)

	err = p.InstallBundle(installOpts)
	require.NoError(t, err)

	// TODO: We can't check this yet because docker driver is printing directly to stdout instead of to the given writer
	// output := p.TestConfig.TestContext.GetOutput()
	// require.Contains(t, output, "Hello World!", "expected action output to contain provided file contents")

	claim, err := p.Claims.Read(p.Manifest.Name)
	require.NoError(t, err, "could not fetch claim")
	require.Equal(t, "Hello World!", claim.Outputs["myfile"], "expected output to match the decoded file contents")
}
