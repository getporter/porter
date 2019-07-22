// +build integration

package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/porter"
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

	installOpts := porter.InstallOptions{}
	installOpts.Insecure = true
	err = installOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	// Install the bundle, assert no error occurs due to Porter home as relative path
	err = p.InstallBundle(installOpts)
	require.NoError(t, err)
}
