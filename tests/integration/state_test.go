//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestState_ParentDirectoryCreation(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Add the test bundle that has a state file with a path requiring parent directory creation
	p.AddTestBundleDir("testdata/bundles/bundle-with-state", false)

	// First, install the bundle - this should succeed as the install action creates the directory
	installOpts := porter.NewInstallOptions()
	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err, "Install should succeed as it creates the parent directory")

	// Now try to upgrade the bundle - this should succeed because our fix creates the parent directory
	upgradeOpts := porter.NewUpgradeOptions()
	err = upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(t, err, "Upgrade should succeed because our fix creates parent directories for state files")
}
