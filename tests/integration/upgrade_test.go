//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestUpgrade_failedInstallation(t *testing.T) {
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.AddTestBundleDir("testdata/bundles/bundle-with-failing-install", false)

	installOpts := porter.NewInstallOptions()
	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.Error(t, err, "Installation should fail")

	upgradeOpts := porter.NewUpgradeOptions()
	err = upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.Error(t, err, "Upgrade should fail, because the installation failed")
}
