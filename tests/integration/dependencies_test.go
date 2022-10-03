//go:build integration

package integration

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependenciesLifecycle(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	namespace := installWordpressBundle(ctx, p)
	defer cleanupWordpressBundle(ctx, p, namespace)

	upgradeWordpressBundle(ctx, p, namespace)

	invokeWordpressBundle(ctx, p, namespace)

	uninstallWordpressBundle(ctx, p, namespace)
}

func publishMySQLBundle(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := ioutil.TempDir("", "porter-mysql")
	require.NoError(p.T(), err, "could not create temp directory to publish the mysql bundle")
	defer os.RemoveAll(bunDir)

	// Rebuild the bundle from a temp directory so that we don't modify the source directory
	// and leave modified files around.
	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/mysql"), bunDir)
	pwd := p.Getwd()
	p.Chdir(bunDir)
	defer p.Chdir(pwd)

	publishOpts := porter.PublishOptions{}
	err = publishOpts.Validate(p.Config)
	require.NoError(p.T(), err, "validation of publish opts for dependent bundle failed")

	err = p.Publish(ctx, publishOpts)
	require.NoError(p.T(), err, "publish of dependent bundle failed")
}

func installWordpressBundle(ctx context.Context, p *porter.TestPorter) (namespace string) {
	// Publish the mysql bundle that we depend upon
	publishMySQLBundle(ctx, p)

	// Install the bundle that has dependencies
	p.CopyDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/wordpress"), ".", false)

	namespace = p.RandomString(10)
	installOpts := porter.NewInstallOptions()
	installOpts.Namespace = namespace
	installOpts.CredentialIdentifiers = []string{"ci"} // Use the ci credential set, porter should remember this for later
	installOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"mysql#namespace=" + namespace,
	}

	// Add a supplemental parameter set to vet dep param resolution
	installOpts.ParameterSets = []string{"myparam"}
	testParamSets := storage.NewParameterSet(namespace, "myparam", secrets.Strategy{
		Name: "mysql#probe-timeout",
		Source: secrets.Source{
			Key:   host.SourceValue,
			Value: "2",
		},
	})

	p.TestParameters.InsertParameterSet(ctx, testParamSets)

	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of install opts for root bundle failed")

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(p.T(), err, "install of root bundle failed namespace %s", namespace)

	// Verify that the dependency claim is present
	i, err := p.Installations.GetInstallation(ctx, namespace, "wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch installation status for the dependency")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being installed successfully")
	c, err := p.Installations.GetLastRun(ctx, namespace, i.Name)
	require.NoError(p.T(), err, "GetLastRun failed")
	resolvedParameters, err := p.Sanitizer.RestoreParameterSet(ctx, c.Parameters, cnab.ExtendedBundle{c.Bundle})
	require.NoError(p.T(), err, "Resolve run failed")
	assert.Equal(p.T(), "porter-ci-mysql", resolvedParameters["mysql-name"], "the dependency param value for 'mysql-name' is incorrect")
	assert.Equal(p.T(), 2, resolvedParameters["probe-timeout"], "the dependency param value for 'probe-timeout' is incorrect")

	// TODO(carolynvs): compare with last parameters set on the installation once we start tracking that

	// Verify that the bundle claim is present
	i, err = p.Installations.GetInstallation(ctx, namespace, "wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being installed successfully")

	return namespace
}

func cleanupWordpressBundle(ctx context.Context, p *porter.TestPorter, namespace string) {
	uninstallOptions := porter.NewUninstallOptions()
	uninstallOptions.Namespace = namespace
	uninstallOptions.CredentialIdentifiers = []string{"ci"}
	uninstallOptions.Delete = true
	err := uninstallOptions.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of uninstall opts for root bundle failed")

	err = p.UninstallBundle(ctx, uninstallOptions)
	require.NoError(p.T(), err, "uninstall of root bundle failed")

	// Verify that the dependency installation is deleted
	i, err := p.Installations.GetInstallation(ctx, namespace, "wordpress-mysql")
	require.ErrorIs(p.T(), err, storage.ErrNotFound{})
	require.Equal(p.T(), storage.Installation{}, i)

	// Verify that the root installation is deleted
	i, err = p.Installations.GetInstallation(ctx, namespace, "wordpress")
	require.ErrorIs(p.T(), err, storage.ErrNotFound{})
	require.Equal(p.T(), storage.Installation{}, i)
}

func upgradeWordpressBundle(ctx context.Context, p *porter.TestPorter, namespace string) {
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

	// Verify that the dependency claim is upgraded
	i, err := p.Installations.GetInstallation(ctx, namespace, "wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")
	c, err := p.Installations.GetLastRun(ctx, i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUpgrade, c.Action, "the dependency wasn't recorded as being upgraded")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being upgraded successfully")

	// Verify that the bundle claim is upgraded
	i, err = p.Installations.GetInstallation(ctx, namespace, "wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	c, err = p.Installations.GetLastRun(ctx, i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUpgrade, c.Action, "the root bundle wasn't recorded as being upgraded")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being upgraded successfully")

	// Check that we are using the original credential set specified during install
	require.Len(p.T(), i.CredentialSets, 1, "expected only one credential set associated to the installation")
	assert.Equal(p.T(), "ci", i.CredentialSets[0], "expected to use the alternate credential set")
}

func invokeWordpressBundle(ctx context.Context, p *porter.TestPorter, namespace string) {
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
	i, err := p.Installations.GetInstallation(ctx, namespace, "wordpress-mysql")
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

func uninstallWordpressBundle(ctx context.Context, p *porter.TestPorter, namespace string) {
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

	// Verify that the dependency claim is uninstalled
	i, err := p.Installations.GetInstallation(ctx, namespace, "wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch installation for the dependency")
	c, err := p.Installations.GetLastRun(ctx, i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUninstall, c.Action, "the dependency wasn't recorded as being uninstalled")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being uninstalled successfully")

	// Verify that the bundle claim is uninstalled
	i, err = p.Installations.GetInstallation(ctx, namespace, "wordpress")
	require.NoError(p.T(), err, "could not fetch installation for the root bundle")
	c, err = p.Installations.GetLastRun(ctx, i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUninstall, c.Action, "the root bundle wasn't recorded as being uninstalled")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being uninstalled successfully")

	// Check that we are now using the original credentials with the bundle
	require.Len(p.T(), i.CredentialSets, 1, "expected only one credential set associated to the installation")
	assert.Equal(p.T(), "ci", i.CredentialSets[0], "expected to use the alternate credential set")

}
