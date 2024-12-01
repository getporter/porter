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

func TestUpgrade_failedInstallation_withForceUpgrade(t *testing.T) {
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
	upgradeOpts.ForceUpgrade = true
	err = upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(t, err, "Upgrade should succeed, because force-upgrade is true")
}

func TestUpgrade_DebugModeAppliesToSingleInvocation(t *testing.T) {
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.AddTestBundleDir("testdata/bundles/bundle-with-custom-action", false)

	installOpts := porter.NewInstallOptions()
	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)

	upgradeOpts := porter.NewUpgradeOptions()
	upgradeOpts.DebugMode = true
	err = upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(t, err)
	output := p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "== Step Template ===")
	p.TestConfig.TestContext.ClearOutputs()

	upgradeOpts = porter.NewUpgradeOptions()
	upgradeOpts.DebugMode = false
	err = upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(t, err)
	output = p.TestConfig.TestContext.GetOutput()
	require.NotContains(t, output, "== Step Template ===")
}
