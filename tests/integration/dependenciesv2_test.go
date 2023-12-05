//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedDependencies(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	ctx := p.SetupIntegrationTest()
	bunDir := setupFS(ctx, p)
	defer os.RemoveAll(bunDir)
	p.Config.SetExperimentalFlags(experimental.FlagDependenciesV2)

	namespace := p.RandomString(10)
	setupMysql(ctx, p, namespace, bunDir)

	setupWordpress_v2(ctx, p, namespace, bunDir)
	upgradeWordpressBundle_v2(ctx, p, namespace)
	invokeWordpressBundle_v2(ctx, p, namespace)
	uninstallWordpressBundle_v2(ctx, p, namespace)
	defer cleanupWordpressBundle_v2(ctx, p, namespace)

}

func setupFS(ctx context.Context, p *porter.TestPorter) string {
	bunDir, err := os.MkdirTemp("", "porter-mysql-")
	require.NoError(p.T(), err, "could not create temp directory at all")

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/wordpressv2"), bunDir+"/wordpress")

	return bunDir
}

func setupMysql(ctx context.Context, p *porter.TestPorter, namespace string, bunDir string) {
	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/mysql"), bunDir+"/mysql")

	p.Chdir(bunDir + "/mysql")

	publishOpts := porter.PublishOptions{}
	publishOpts.Force = true
	err := publishOpts.Validate(p.Config)
	require.NoError(p.T(), err, "validation of publish opts for dependent bundle failed")

	err = p.Publish(ctx, publishOpts)
	require.NoError(p.T(), err, "publish of dependent bundle failed")
	installOpts := porter.NewInstallOptions()

	installOpts.Namespace = namespace
	installOpts.CredentialIdentifiers = []string{"ci"} // Use the ci credential set, porter should remember this for later

	err = installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of install opts for shared mysql bundle failed")

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(p.T(), err, "install of shared mysql bundle failed namespace %s", namespace)

	mysqlinst, err := p.Installations.GetInstallation(ctx, namespace, "mysql")
	require.NoError(p.T(), err, "could not fetch installation status for the dependency")

	//Set the label on the installaiton so Porter knows to grab it
	mysqlinst.SetLabel("sh.porter.SharingGroup", "myapp")
	err = p.Installations.UpdateInstallation(ctx, mysqlinst)
	require.NoError(p.T(), err, "could not add label to mysql inst")

}

func setupWordpress_v2(ctx context.Context, p *porter.TestPorter, namespace string, bunDir string) {

	p.Chdir(bunDir + "/wordpress")

	publishOpts := porter.PublishOptions{}
	publishOpts.Force = true
	err := publishOpts.Validate(p.Config)
	require.NoError(p.T(), err, "validation of publish opts for dependent bundle failed")

	err = p.Publish(ctx, publishOpts)
	require.NoError(p.T(), err, "publish of dependent bundle failed")

	installOpts := porter.NewInstallOptions()
	installOpts.Namespace = namespace
	installOpts.CredentialIdentifiers = []string{"ci"} // Use the ci credential set, porter should remember this for later
	installOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
	}

	err = installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of install opts for root bundle failed")

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(p.T(), err, "install of root bundle failed namespace %s", namespace)

	numInst, err := p.Installations.ListInstallations(ctx, storage.ListOptions{Namespace: namespace})
	assert.Equal(p.T(), len(numInst), 2)

	i, err := p.Installations.GetInstallation(ctx, namespace, "mysql")
	require.NoError(p.T(), err, "could not fetch installation status for the dependency")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being installed successfully")

	// Verify that the bundle claim is present
	i, err = p.Installations.GetInstallation(ctx, namespace, "wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being installed successfully")
}

func cleanupWordpressBundle_v2(ctx context.Context, p *porter.TestPorter, namespace string) {
	uninstallOptions := porter.NewUninstallOptions()
	uninstallOptions.Namespace = namespace
	uninstallOptions.CredentialIdentifiers = []string{"ci"}
	uninstallOptions.Delete = true
	err := uninstallOptions.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of uninstall opts for root bundle failed")

	err = p.UninstallBundle(ctx, uninstallOptions)
	require.NoError(p.T(), err, "uninstall of root bundle failed")

	// This shouldn't get deleted, it existed before, and it should exist after
	i, err := p.Installations.GetInstallation(ctx, namespace, "mysql")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being installed successfully")

	// Verify that the root installation is deleted
	i, err = p.Installations.GetInstallation(ctx, namespace, "wordpress")
	require.ErrorIs(p.T(), err, storage.ErrNotFound{})
	require.Equal(p.T(), storage.Installation{}, i)
}

func upgradeWordpressBundle_v2(ctx context.Context, p *porter.TestPorter, namespace string) {
	upgradeOpts := porter.NewUpgradeOptions()
	upgradeOpts.Namespace = namespace
	// do not specify credential sets, porter should reuse what was specified from install
	upgradeOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"mysql#namespace=" + namespace,
	}
	err := upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of upgrade opts for root bundle failed")

	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(p.T(), err, "upgrade of root bundle failed")

	// Verify that the dependency claim is still installed
	// upgrade should not change our status
	i, err := p.Installations.GetInstallation(ctx, namespace, "mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")

	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being upgraded successfully")

	// Verify that the bundle claim is upgraded
	i, err = p.Installations.GetInstallation(ctx, namespace, "wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	c, err := p.Installations.GetLastRun(ctx, i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUpgrade, c.Action, "the root bundle wasn't recorded as being upgraded")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being upgraded successfully")

	// Check that we are using the original credential set specified during install
	require.Len(p.T(), i.CredentialSets, 1, "expected only one credential set associated to the installation")
	assert.Equal(p.T(), "ci", i.CredentialSets[0], "expected to use the alternate credential set")
}

func invokeWordpressBundle_v2(ctx context.Context, p *porter.TestPorter, namespace string) {
	invokeOpts := porter.NewInvokeOptions()
	invokeOpts.Namespace = namespace
	invokeOpts.Action = "ping"
	// Use a different set of creds to run this rando command
	invokeOpts.CredentialIdentifiers = []string{"ci2"}
	invokeOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
	}
	err := invokeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of invoke opts for root bundle failed")

	err = p.InvokeBundle(ctx, invokeOpts)
	require.NoError(p.T(), err, "invoke of root bundle failed")

	// Verify that the dependency claim is invoked

	i, err := p.Installations.GetInstallation(ctx, namespace, "mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")
	c, err := p.Installations.GetLastRun(ctx, i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), "ping", c.Action, "the dependency wasn't recorded as being invoked")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being invoked successfully")

	// Verify that the bundle claim is invoked
	i, err = p.Installations.GetInstallation(ctx, namespace, "wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	c, err = p.Installations.GetLastRun(ctx, i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), "ping", c.Action, "the root bundle wasn't recorded as being invoked")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being invoked successfully")

	// Check that we are now using the alternate credentials with the bundle
	require.Len(p.T(), i.CredentialSets, 1, "expected only one credential set associated to the installation")
	assert.Equal(p.T(), "ci2", i.CredentialSets[0], "expected to use the alternate credential set")
}

func uninstallWordpressBundle_v2(ctx context.Context, p *porter.TestPorter, namespace string) {

	uninstallOptions := porter.NewUninstallOptions()
	uninstallOptions.Namespace = namespace
	// Now go back to using the original set of credentials
	uninstallOptions.CredentialIdentifiers = []string{"ci"}
	uninstallOptions.Params = []string{
		"namespace=" + namespace,
		"mysql#namespace=" + namespace,
	}
	err := uninstallOptions.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of uninstall opts for root bundle failed")

	err = p.UninstallBundle(ctx, uninstallOptions)
	require.NoError(p.T(), err, "uninstall of root bundle failed")

	// Verify that the bundle claim is uninstalled
	i, err := p.Installations.GetInstallation(ctx, namespace, "wordpress")
	require.NoError(p.T(), err, "could not fetch installation for the root bundle")
	c, err := p.Installations.GetLastRun(ctx, i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUninstall, c.Action, "the root bundle wasn't recorded as being uninstalled")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being uninstalled successfully")

	// Check that we are now using the original credentials with the bundle
	require.Len(p.T(), i.CredentialSets, 1, "expected only one credential set associated to the installation")
	assert.Equal(p.T(), "ci", i.CredentialSets[0], "expected to use the alternate credential set")

}
