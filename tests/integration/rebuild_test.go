//go:build integration
// +build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/manifest"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	yaml "get.porter.sh/porter/pkg/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebuild_InstallNewBundle(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = false

	// Create a bundle
	err := p.Create()
	require.NoError(t, err)

	// Install a bundle without building first
	installOpts := porter.NewInstallOptions()
	err = installOpts.Validate([]string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(context.Background(), installOpts)
	assert.NoError(t, err, "install should have succeeded")
}

func TestRebuild_UpgradeModifiedBundle(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = false

	// Install a bundle
	err := p.Create()
	require.NoError(t, err)
	installOpts := porter.NewInstallOptions()
	err = installOpts.Validate([]string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(context.Background(), installOpts)
	require.NoError(t, err)

	// Modify the porter.yaml to trigger a rebuild
	m, err := manifest.ReadManifest(p.Context, config.Name)
	require.NoError(t, err)
	m.Version = "0.2.0"
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	err = p.FileSystem.WriteFile(config.Name, data, pkg.FileModeWritable)
	require.NoError(t, err)

	// Upgrade the bundle
	upgradeOpts := porter.NewUpgradeOptions()
	err = upgradeOpts.Validate([]string{}, p.Porter)
	require.NoError(t, err)
	err = p.UpgradeBundle(context.Background(), upgradeOpts)
	require.NoError(t, err, "upgrade should have succeeded")

	gotOutput := p.TestConfig.TestContext.GetOutput()
	buildCount := strings.Count(gotOutput, "Building bundle ===>")
	assert.Equal(t, 2, buildCount, "expected a rebuild before upgrade")

	// Verify that the bundle's version matches the updated version in the porter.yaml
	// TODO: separate ListBundle's printing from fetching claims
}

func TestRebuild_GenerateCredentialsNewBundle(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = false

	// Create a bundle that uses credentials
	p.AddTestBundleDir("testdata/bundles/bundle-with-credentials", true)

	credentialOptions := porter.CredentialOptions{}
	credentialOptions.Silent = true
	err := credentialOptions.Validate([]string{}, p.Porter)
	require.NoError(t, err)
	err = p.GenerateCredentials(context.Background(), credentialOptions)
	assert.NoError(t, err)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, gotOutput, "Building bundle ===>", "expected a rebuild before generating credentials")
}

func TestRebuild_GenerateCredentialsExistingBundle(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = false

	// Create a bundle that uses credentials
	p.AddTestBundleDir("testdata/bundles/bundle-with-credentials", true)

	credentialOptions := porter.CredentialOptions{}
	credentialOptions.Silent = true
	err := credentialOptions.Validate([]string{}, p.Porter)
	require.NoError(t, err)
	err = p.GenerateCredentials(context.Background(), credentialOptions)
	require.NoError(t, err)

	// Modify the porter.yaml to trigger a rebuild
	m, err := manifest.ReadManifest(p.Context, config.Name)
	require.NoError(t, err)
	m.Version = "0.2.0"
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	err = p.FileSystem.WriteFile(config.Name, data, pkg.FileModeWritable)
	require.NoError(t, err)

	// Re-generate the credentials
	err = p.GenerateCredentials(context.Background(), credentialOptions)
	require.NoError(t, err)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	buildCount := strings.Count(gotOutput, "Building bundle ===>")
	assert.Equal(t, 2, buildCount, "expected a rebuild before generating credentials")
}
