//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistentParameters(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Feature flag must be set before AddTestBundleDir since that validates the manifest
	p.Config.SetExperimentalFlags(experimental.FlagPersistentParameters)

	bundleName := p.AddTestBundleDir("testdata/bundles/bundle-with-persistent-params", true)

	// Install: provide the persistent parameter value
	installOpts := porter.NewInstallOptions()
	installOpts.Params = []string{"resource-group=my-rg"}
	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err, "install failed")

	// Verify the parameter was captured as an output
	outputs, err := p.Installations.GetLastOutputs(ctx, "", bundleName)
	require.NoError(t, err)
	rgOutput, ok := outputs.GetByName("resource-group")
	require.True(t, ok, "resource-group output should be persisted after install")
	assert.Equal(t, "my-rg", string(rgOutput.Value))

	// Upgrade: do NOT provide the parameter — it must come from the persisted output
	upgradeOpts := porter.NewUpgradeOptions()
	err = upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(t, err, "upgrade failed: persistent parameter was not carried over from install")

	// Verify the output is still the same after upgrade
	outputs, err = p.Installations.GetLastOutputs(ctx, "", bundleName)
	require.NoError(t, err)
	rgOutput, ok = outputs.GetByName("resource-group")
	require.True(t, ok, "resource-group output should be persisted after upgrade")
	assert.Equal(t, "my-rg", string(rgOutput.Value), "persistent parameter value should match install value")
}
