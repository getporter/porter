//go:build integration
// +build integration

package integration

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependenciesLifecycle(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = true

	namespace := installWordpressBundle(p)
	defer cleanupWordpressBundle(p, namespace)

	upgradeWordpressBundle(p, namespace)

	invokeWordpressBundle(p, namespace)

	uninstallWordpressBundle(p, namespace)
}

func publishMySQLBundle(p *porter.TestPorter) {
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
	err = publishOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for dependent bundle failed")

	err = p.Publish(context.Background(), publishOpts)
	require.NoError(p.T(), err, "publish of dependent bundle failed")
}

func installWordpressBundle(p *porter.TestPorter) (namespace string) {
	// Publish the mysql bundle that we depend upon
	publishMySQLBundle(p)

	// Install the bundle that has dependencies
	p.CopyDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/wordpress"), ".", false)

	namespace = p.RandomString(10)
	installOpts := porter.NewInstallOptions()
	installOpts.CredentialIdentifiers = []string{"ci"}
	installOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"mysql#namespace=" + namespace,
	}

	// Add a supplemental parameter set to vet dep param resolution
	testParamSets := parameters.NewParameterSet(namespace, "myparam", secrets.Strategy{
		Name: "mysql#probe-timeout",
		Source: secrets.Source{
			Key:   secrets.SourceSecret,
			Value: "2",
		},
	})

	err := p.Parameters.UpsertParameterSet(testParamSets)
	require.NoError(p.T(), err, "could not create a parameter set")

	err = installOpts.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of install opts for root bundle failed")

	err = p.InstallBundle(context.Background(), installOpts)
	require.NoError(p.T(), err, "install of root bundle failed")

	// Verify that the dependency claim is present
	i, err := p.Claims.GetInstallation("", "wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch installation status for the dependency")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being installed successfully")
	c, err := p.Claims.GetLastRun("", i.Name)
	require.NoError(p.T(), err, "GetLastRun failed")
	assert.Equal(p.T(), "porter-ci-mysql", c.Parameters["mysql-name"], "the dependency param value for 'mysql-name' is incorrect")
	assert.Equal(p.T(), float64(2), c.Parameters["probe-timeout"], "the dependency param value for 'probe-timeout' is incorrect")

	// TODO(carolynvs): compare with last parameters set on the installation once we start tracking that

	// Verify that the bundle claim is present
	i, err = p.Claims.GetInstallation("", "wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being installed successfully")

	return namespace
}

func cleanupWordpressBundle(p *porter.TestPorter, namespace string) {
	uninstallOptions := porter.NewUninstallOptions()
	uninstallOptions.CredentialIdentifiers = []string{"ci"}
	uninstallOptions.Delete = true
	err := uninstallOptions.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of uninstall opts for root bundle failed")

	err = p.UninstallBundle(context.Background(), uninstallOptions)
	require.NoError(p.T(), err, "uninstall of root bundle failed")

	// Verify that the dependency installation is deleted
	i, err := p.Claims.GetInstallation("", "wordpress-mysql")
	require.ErrorIs(p.T(), err, storage.ErrNotFound{})
	require.Equal(p.T(), claims.Installation{}, i)

	// Verify that the root installation is deleted
	i, err = p.Claims.GetInstallation("", "wordpress")
	require.ErrorIs(p.T(), err, storage.ErrNotFound{})
	require.Equal(p.T(), claims.Installation{}, i)
}

func upgradeWordpressBundle(p *porter.TestPorter, namespace string) {
	upgradeOpts := porter.NewUpgradeOptions()
	upgradeOpts.CredentialIdentifiers = []string{"ci"}
	upgradeOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"mysql#namespace=" + namespace,
	}
	err := upgradeOpts.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of upgrade opts for root bundle failed")

	err = p.UpgradeBundle(context.Background(), upgradeOpts)
	require.NoError(p.T(), err, "upgrade of root bundle failed")

	// Verify that the dependency claim is upgraded
	i, err := p.Claims.GetInstallation("", "wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")
	c, err := p.Claims.GetLastRun(i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUpgrade, c.Action, "the dependency wasn't recorded as being upgraded")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being upgraded successfully")

	// Verify that the bundle claim is upgraded
	i, err = p.Claims.GetInstallation("", "wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	c, err = p.Claims.GetLastRun(i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUpgrade, c.Action, "the root bundle wasn't recorded as being upgraded")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being upgraded successfully")
}

func invokeWordpressBundle(p *porter.TestPorter, namespace string) {
	invokeOpts := porter.NewInvokeOptions()
	invokeOpts.Action = "ping"
	invokeOpts.CredentialIdentifiers = []string{"ci"}
	invokeOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
	}
	err := invokeOpts.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of invoke opts for root bundle failed")

	err = p.InvokeBundle(context.Background(), invokeOpts)
	require.NoError(p.T(), err, "invoke of root bundle failed")

	// Verify that the dependency claim is invoked
	i, err := p.Claims.GetInstallation("", "wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")
	c, err := p.Claims.GetLastRun(i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), "ping", c.Action, "the dependency wasn't recorded as being invoked")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being invoked successfully")

	// Verify that the bundle claim is invoked
	i, err = p.Claims.GetInstallation("", "wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	c, err = p.Claims.GetLastRun(i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), "ping", c.Action, "the root bundle wasn't recorded as being invoked")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being invoked successfully")
}

func uninstallWordpressBundle(p *porter.TestPorter, namespace string) {
	uninstallOptions := porter.NewUninstallOptions()
	uninstallOptions.CredentialIdentifiers = []string{"ci"}
	uninstallOptions.Params = []string{
		"namespace=" + namespace,
		"mysql#namespace=" + namespace,
	}
	err := uninstallOptions.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of uninstall opts for root bundle failed")

	err = p.UninstallBundle(context.Background(), uninstallOptions)
	require.NoError(p.T(), err, "uninstall of root bundle failed")

	// Verify that the dependency claim is uninstalled
	i, err := p.Claims.GetInstallation("", "wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch installation for the dependency")
	c, err := p.Claims.GetLastRun(i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUninstall, c.Action, "the dependency wasn't recorded as being uninstalled")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the dependency wasn't recorded as being uninstalled successfully")

	// Verify that the bundle claim is uninstalled
	i, err = p.Claims.GetInstallation("", "wordpress")
	require.NoError(p.T(), err, "could not fetch installation for the root bundle")
	c, err = p.Claims.GetLastRun(i.Namespace, i.Name)
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), cnab.ActionUninstall, c.Action, "the root bundle wasn't recorded as being uninstalled")
	assert.Equal(p.T(), cnab.StatusSucceeded, i.Status.ResultStatus, "the root bundle wasn't recorded as being uninstalled successfully")
}
