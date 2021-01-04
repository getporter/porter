// +build integration

package porter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallFromTag_ManageFromClaim(t *testing.T) {
	const tag = "getporter/porter-hello:v0.1.1"

	p := NewTestPorter(t)
	p.CacheTestBundle("../../examples/hello", tag)

	installOpts := NewInstallOptions()
	installOpts.Name = "hello"
	installOpts.Tag = tag
	err := installOpts.Validate(nil, p.Porter)
	require.NoError(t, err, "InstallOptions.Validate failed")

	err = p.InstallBundle(installOpts)
	require.NoError(t, err, "InstallBundle failed")

	upgradeOpts := NewUpgradeOptions()
	upgradeOpts.Name = installOpts.Name
	err = upgradeOpts.Validate(nil, p.Porter)

	err = p.UpgradeBundle(upgradeOpts)
	require.NoError(t, err, "UpgradeBundle failed")

	uninstallOpts := NewUninstallOptions()
	uninstallOpts.Name = installOpts.Name
	err = uninstallOpts.Validate(nil, p.Porter)

	err = p.UninstallBundle(uninstallOpts)
	require.NoError(t, err, "UninstallBundle failed")
}
