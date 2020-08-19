// +build integration

package tests

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/porter"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependenciesLifecycle(t *testing.T) {
	t.Skip("TODO: Implement parameter sources #1069")

	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	namespace := installWordpressBundle(p)
	defer cleanupWordpressBundle(p, namespace)

	upgradeWordpressBundle(p, namespace)

	invokeWordpressBundle(p, namespace)

	uninstallWordpressBundle(p, namespace)
}

func randomString(len int) string {
	rand.Seed(time.Now().UnixNano())
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		//A=97 and Z = 97+25
		bytes[i] = byte(97 + rand.Intn(25))
	}
	return string(bytes)
}

func publishMySQLBundle(p *porter.TestPorter) {
	mysqlBundlePath := filepath.Join(p.TestDir, "../build/testdata/bundles/mysql")
	err := os.Chdir(mysqlBundlePath)
	require.NoError(p.T(), err, "could not change into the test mysql bundle directory")
	defer os.Chdir(p.BundleDir)

	publishOpts := porter.PublishOptions{}
	err = publishOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for dependent bundle failed")

	err = p.Publish(publishOpts)
	require.NoError(p.T(), err, "publish of dependent bundle failed")
}

func installWordpressBundle(p *porter.TestPorter) (namespace string) {
	// Publish the mysql bundle that we depend upon
	publishMySQLBundle(p)

	// Install the bundle that has dependencies
	p.CopyDirectory(filepath.Join(p.TestDir, "../build/testdata/bundles/wordpress"), ".", false)

	namespace = randomString(10)
	installOpts := porter.InstallOptions{}
	installOpts.CredentialIdentifiers = []string{"ci"}
	installOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#namespace=" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
	}
	// Add a supplemental parameter set to vet dep param resolution
	installOpts.ParameterSets = []string{filepath.Join(p.TestDir, "testdata/parameter-set-for-dependencies.json")}

	err := installOpts.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of install opts for root bundle failed")

	err = p.InstallBundle(installOpts)
	require.NoError(p.T(), err, "install of root bundle failed")

	// Verify that the dependency claim is present
	i, err := p.Claims.ReadInstallationStatus("wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch installation status for the dependency")
	assert.Equal(p.T(), claim.StatusSucceeded, i.GetLastStatus(), "the dependency wasn't recorded as being installed successfully")
	c, err := i.GetLastClaim()
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), "porter-ci-mysql-"+namespace, c.Parameters["mysql-name"], "the dependency param value for 'mysql-name' is incorrect")
	assert.Equal(p.T(), "mydb", c.Parameters["database-name"], "the dependency param value for 'dabaase-name' is incorrect")

	// Verify that the bundle claim is present
	i, err = p.Claims.ReadInstallationStatus("wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	assert.Equal(p.T(), claim.StatusSucceeded, i.GetLastStatus(), "the root bundle wasn't recorded as being installed successfully")

	return namespace
}

func cleanupWordpressBundle(p *porter.TestPorter, namespace string) {
	uninstallOptions := porter.UninstallOptions{}
	uninstallOptions.CredentialIdentifiers = []string{"ci"}
	uninstallOptions.Delete = true
	uninstallOptions.Params = []string{
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
	}
	err := uninstallOptions.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of uninstall opts for root bundle failed")

	err = p.UninstallBundle(uninstallOptions)
	require.NoError(p.T(), err, "uninstall of root bundle failed")

	// Verify that the dependency installation is deleted
	i, err := p.Claims.ReadInstallation("wordpress-mysql")
	require.EqualError(p.T(), err, "Installation does not exist")
	require.Equal(p.T(), claim.Installation{}, i)

	// Verify that the root installation is deleted
	i, err = p.Claims.ReadInstallation("wordpress")
	require.EqualError(p.T(), err, "Installation does not exist")
	require.Equal(p.T(), claim.Installation{}, i)
}

func upgradeWordpressBundle(p *porter.TestPorter, namespace string) {
	upgradeOpts := porter.UpgradeOptions{}
	upgradeOpts.CredentialIdentifiers = []string{"ci"}
	upgradeOpts.Params = []string{ // See https://github.com/deislabs/porter/issues/474
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#namespace=" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
	}
	err := upgradeOpts.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of upgrade opts for root bundle failed")

	err = p.UpgradeBundle(upgradeOpts)
	require.NoError(p.T(), err, "upgrade of root bundle failed")

	// Verify that the dependency claim is upgraded
	i, err := p.Claims.ReadInstallationStatus("wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")
	c, err := i.GetLastClaim()
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), claim.ActionUpgrade, c.Action, "the dependency wasn't recorded as being upgraded")
	assert.Equal(p.T(), claim.StatusSucceeded, i.GetLastStatus(), "the dependency wasn't recorded as being upgraded successfully")

	// Verify that the bundle claim is upgraded
	i, err = p.Claims.ReadInstallationStatus("wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	c, err = i.GetLastClaim()
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), claim.ActionUpgrade, c.Action, "the root bundle wasn't recorded as being upgraded")
	assert.Equal(p.T(), claim.StatusSucceeded, i.GetLastStatus(), "the root bundle wasn't recorded as being upgraded successfully")
}

func invokeWordpressBundle(p *porter.TestPorter, namespace string) {
	invokeOpts := porter.InvokeOptions{Action: "ping"}
	invokeOpts.CredentialIdentifiers = []string{"ci"}
	invokeOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#namespace=" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
	}
	err := invokeOpts.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of invoke opts for root bundle failed")

	err = p.InvokeBundle(invokeOpts)
	require.NoError(p.T(), err, "invoke of root bundle failed")

	// Verify that the dependency claim is invoked
	i, err := p.Claims.ReadInstallationStatus("wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")
	c, err := i.GetLastClaim()
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), "ping", c.Action, "the dependency wasn't recorded as being invoked")
	assert.Equal(p.T(), claim.StatusSucceeded, i.GetLastStatus(), "the dependency wasn't recorded as being invoked successfully")

	// Verify that the bundle claim is invoked
	i, err = p.Claims.ReadInstallationStatus("wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	c, err = i.GetLastClaim()
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), "ping", c.Action, "the root bundle wasn't recorded as being invoked")
	assert.Equal(p.T(), claim.StatusSucceeded, i.GetLastStatus(), "the root bundle wasn't recorded as being invoked successfully")
}

func uninstallWordpressBundle(p *porter.TestPorter, namespace string) {
	uninstallOptions := porter.UninstallOptions{}
	uninstallOptions.CredentialIdentifiers = []string{"ci"}
	uninstallOptions.Params = []string{
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
	}
	err := uninstallOptions.Validate([]string{}, p.Porter)
	require.NoError(p.T(), err, "validation of uninstall opts for root bundle failed")

	err = p.UninstallBundle(uninstallOptions)
	require.NoError(p.T(), err, "uninstall of root bundle failed")

	// Verify that the dependency claim is uninstalled
	i, err := p.Claims.ReadInstallationStatus("wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch installation for the dependency")
	c, err := i.GetLastClaim()
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), claim.ActionUninstall, c.Action, "the dependency wasn't recorded as being uninstalled")
	assert.Equal(p.T(), claim.StatusSucceeded, i.GetLastStatus(), "the dependency wasn't recorded as being uninstalled successfully")

	// Verify that the bundle claim is uninstalled
	i, err = p.Claims.ReadInstallationStatus("wordpress")
	require.NoError(p.T(), err, "could not fetch installation for the root bundle")
	c, err = i.GetLastClaim()
	require.NoError(p.T(), err, "GetLastClaim failed")
	assert.Equal(p.T(), claim.ActionUninstall, c.Action, "the root bundle wasn't recorded as being uninstalled")
	assert.Equal(p.T(), claim.StatusSucceeded, i.GetLastStatus(), "the root bundle wasn't recorded as being uninstalled successfully")
}
