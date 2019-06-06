// +build integration

package tests

import (
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestRebuild_InstallNewBundle(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	// Create a bundle
	err := p.Create()
	require.NoError(t, err)

	// Install a bundle without building first
	installOpts := porter.InstallOptions{}
	installOpts.Insecure = true
	installOpts.Validate([]string{}, p.Context)
	err = p.InstallBundle(installOpts)
	assert.NoError(t, err, "install should have succeeded")
}

func TestRebuild_UpgradeModifiedBundle(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	// Install a bundle
	err := p.Create()
	require.NoError(t, err)
	installOpts := porter.InstallOptions{}
	installOpts.Insecure = true
	installOpts.Validate([]string{}, p.Context)
	err = p.InstallBundle(installOpts)
	require.NoError(t, err)

	// Modify the porter.yaml to trigger a rebuild
	m, err := p.ReadManifest(config.Name)
	require.NoError(t, err)
	m.Version = "0.2.0"
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	err = p.FileSystem.WriteFile(config.Name, data, 0644)
	require.NoError(t, err)

	// Upgrade the bundle
	upgradeOpts := porter.UpgradeOptions{}
	upgradeOpts.Insecure = true
	upgradeOpts.Validate([]string{}, p.Context)
	err = p.UpgradeBundle(upgradeOpts)
	require.NoError(t, err, "upgrade should have succeeded")

	gotOutput := p.TestConfig.TestContext.GetOutput()
	buildCount := strings.Count(gotOutput, "Building bundle ===>")
	assert.Equal(t, 2, buildCount, "expected a rebuild before upgrade")

	// Verify that the bundle's version matches the updated version in the porter.yaml
	// TODO: separate ListBundle's printing from fetching claims
}
