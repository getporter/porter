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

// TestUpgrade_paramSetOverridesPreviousParam verifies that a parameter set
// specified on upgrade takes precedence over a parameter value persisted on
// the installation from a prior --param flag.
func TestUpgrade_paramSetOverridesPreviousParam(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.AddTestBundleDir("testdata/bundles/bundle-with-param-set-upgrade", false)

	// Install with an explicit --param so the value is persisted on the installation.
	installOpts := porter.NewInstallOptions()
	installOpts.Params = []string{"name=old-value"}
	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "install: old-value", "sanity check: install used the --param value")
	p.TestConfig.TestContext.ClearOutputs()

	// Create a parameter set with a different value for 'name'.
	err = p.TestParameters.InsertParameterSet(ctx, storage.NewParameterSet("", "upgrade-params",
		secrets.SourceMap{
			Name: "name",
			Source: secrets.Source{
				Strategy: host.SourceValue,
				Hint:     "paramset-value",
			},
		},
	))
	require.NoError(t, err)

	// Upgrade using the parameter set only — no --param flag.
	upgradeOpts := porter.NewUpgradeOptions()
	upgradeOpts.ParameterSets = []string{"upgrade-params"}
	err = upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(t, err)

	output = p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "upgrade: paramset-value",
		"parameter set value must win over the persisted --param value from install")
	require.NotContains(t, output, "old-value",
		"old persisted --param value must not be used when a parameter set is present")
}

// TestUpgrade_cliParamOverridesParamSet verifies that an explicit --param flag
// on upgrade still takes precedence over a parameter set value.
func TestUpgrade_cliParamOverridesParamSet(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.AddTestBundleDir("testdata/bundles/bundle-with-param-set-upgrade", false)

	installOpts := porter.NewInstallOptions()
	installOpts.Params = []string{"name=initial-value"}
	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)
	p.TestConfig.TestContext.ClearOutputs()

	err = p.TestParameters.InsertParameterSet(ctx, storage.NewParameterSet("", "upgrade-params",
		secrets.SourceMap{
			Name: "name",
			Source: secrets.Source{
				Strategy: host.SourceValue,
				Hint:     "paramset-value",
			},
		},
	))
	require.NoError(t, err)

	// Upgrade with both a parameter set and an explicit --param; --param must win.
	upgradeOpts := porter.NewUpgradeOptions()
	upgradeOpts.ParameterSets = []string{"upgrade-params"}
	upgradeOpts.Params = []string{"name=cli-override"}
	err = upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "upgrade: cli-override",
		"explicit --param must take precedence over parameter set value")
}
