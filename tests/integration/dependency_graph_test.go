//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDependencyGraph_CrossDependencyOutputs tests that dependencies can reference
// outputs from other dependencies (not just the parent bundle)
func TestDependencyGraph_CrossDependencyOutputs(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	namespace := p.RandomString(10)

	// Publish the database bundle
	publishDatabaseBundle(ctx, p)

	// Publish the app bundle
	publishAppBundle(ctx, p)

	// Install the parent bundle which has cross-dependency output references
	installParentBundle(ctx, p, namespace)
	defer cleanupParentBundle(ctx, p, namespace)

	// Upgrade the parent bundle to test that outputs are resolved during upgrade
	upgradeParentBundle(ctx, p, namespace)

	// Uninstall the parent bundle
	uninstallParentBundle(ctx, p, namespace)
}

func publishDatabaseBundle(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-dep-graph-db")
	require.NoError(p.T(), err, "could not create temp directory for database bundle")
	defer os.RemoveAll(bunDir)

	// Copy the bundle to a temp directory
	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/dep-graph-db"), bunDir)
	pwd := p.Getwd()
	p.Chdir(bunDir)
	defer p.Chdir(pwd)

	// Publish the bundle
	publishOpts := porter.PublishOptions{}
	publishOpts.Force = true
	err = publishOpts.Validate(p.Config)
	require.NoError(p.T(), err, "validation of publish opts for database bundle failed")

	err = p.Publish(ctx, publishOpts)
	require.NoError(p.T(), err, "publish of database bundle failed")
}

func publishAppBundle(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-dep-graph-app")
	require.NoError(p.T(), err, "could not create temp directory for app bundle")
	defer os.RemoveAll(bunDir)

	// Copy the bundle to a temp directory
	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/dep-graph-app"), bunDir)
	pwd := p.Getwd()
	p.Chdir(bunDir)
	defer p.Chdir(pwd)

	// Publish the bundle
	publishOpts := porter.PublishOptions{}
	publishOpts.Force = true
	err = publishOpts.Validate(p.Config)
	require.NoError(p.T(), err, "validation of publish opts for app bundle failed")

	err = p.Publish(ctx, publishOpts)
	require.NoError(p.T(), err, "publish of app bundle failed")
}

func installParentBundle(ctx context.Context, p *porter.TestPorter, namespace string) {
	// Copy the parent bundle
	p.CopyDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/dep-graph-parent"), ".", false)

	installOpts := porter.NewInstallOptions()
	installOpts.Namespace = namespace

	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of install opts for parent bundle failed")

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(p.T(), err, "install of parent bundle failed")

	// Verify that the database dependency was installed
	dbInstallation, err := p.Installations.GetInstallation(ctx, namespace, "dep-graph-parent-database")
	require.NoError(p.T(), err, "could not fetch installation for database dependency")
	assert.Equal(p.T(), "succeeded", string(dbInstallation.Status.ResultStatus), "database dependency should be installed successfully")

	// Verify that the database output was created
	dbOutput, err := p.Installations.GetLastOutput(ctx, namespace, "dep-graph-parent-database", "connstr")
	require.NoError(p.T(), err, "could not fetch database output")
	assert.Equal(p.T(), "db://localhost:5432/testdb", string(dbOutput.Value), "database connection string should match")

	// Verify that the app dependency was installed
	appInstallation, err := p.Installations.GetInstallation(ctx, namespace, "dep-graph-parent-app")
	require.NoError(p.T(), err, "could not fetch installation for app dependency")
	assert.Equal(p.T(), "succeeded", string(appInstallation.Status.ResultStatus), "app dependency should be installed successfully")

	// Verify that the parent installation was created
	parentInstallation, err := p.Installations.GetInstallation(ctx, namespace, "dep-graph-parent")
	require.NoError(p.T(), err, "could not fetch installation for parent bundle")
	assert.Equal(p.T(), "succeeded", string(parentInstallation.Status.ResultStatus), "parent bundle should be installed successfully")

	// The app bundle validates that it received the correct connection string in its install step
	// If the app installed successfully (which we already verified above), the parameter was resolved correctly
}

func upgradeParentBundle(ctx context.Context, p *porter.TestPorter, namespace string) {
	upgradeOpts := porter.NewUpgradeOptions()
	upgradeOpts.Namespace = namespace

	err := upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of upgrade opts for parent bundle failed")

	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(p.T(), err, "upgrade of parent bundle failed")

	// Verify that the database output was updated
	dbOutput, err := p.Installations.GetLastOutput(ctx, namespace, "dep-graph-parent-database", "connstr")
	require.NoError(p.T(), err, "could not fetch database output after upgrade")
	assert.Equal(p.T(), "db://localhost:5432/testdb-upgraded", string(dbOutput.Value), "database connection string should be upgraded")

	// Verify that the app was upgraded successfully
	// The app validates the connection string in its upgrade step
	appInstallation, err := p.Installations.GetInstallation(ctx, namespace, "dep-graph-parent-app")
	require.NoError(p.T(), err, "could not fetch installation for app dependency after upgrade")
	assert.Equal(p.T(), "succeeded", string(appInstallation.Status.ResultStatus), "app should have validated the upgraded connection string")
}

func uninstallParentBundle(ctx context.Context, p *porter.TestPorter, namespace string) {
	uninstallOpts := porter.NewUninstallOptions()
	uninstallOpts.Namespace = namespace

	err := uninstallOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of uninstall opts for parent bundle failed")

	err = p.UninstallBundle(ctx, uninstallOpts)
	require.NoError(p.T(), err, "uninstall of parent bundle failed")
}

func cleanupParentBundle(ctx context.Context, p *porter.TestPorter, namespace string) {
	// Delete the parent installation
	p.Installations.RemoveInstallation(ctx, namespace, "dep-graph-parent")

	// Delete the dependency installations
	p.Installations.RemoveInstallation(ctx, namespace, "dep-graph-parent-database")
	p.Installations.RemoveInstallation(ctx, namespace, "dep-graph-parent-app")
}
