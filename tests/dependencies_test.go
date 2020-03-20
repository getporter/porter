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
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	namespace := installWordpressBundle(p)
	defer cleanupWordpressBundle(p)

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
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
		// TODO: Remove this once we put back in dependency parameter defaults in #800
		"mysql#mysql-user=wordpress",
	}
	err := installOpts.Validate([]string{}, p.Context)
	require.NoError(p.T(), err, "validation of install opts for root bundle failed")

	err = p.InstallBundle(installOpts)
	require.NoError(p.T(), err, "install of root bundle failed")

	// Verify that the dependency claim is present
	c, err := p.Claims.Read("wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")
	assert.Equal(p.T(), claim.StatusSuccess, c.Result.Status, "the dependency wasn't recorded as being installed successfully")

	// Verify that the bundle claim is present
	c, err = p.Claims.Read("wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	assert.Equal(p.T(), claim.StatusSuccess, c.Result.Status, "the root bundle wasn't recorded as being installed successfully")

	return namespace
}

func cleanupWordpressBundle(p *porter.TestPorter) {
	uninstallOpts := porter.UninstallOptions{}
	uninstallOpts.CredentialIdentifiers = []string{"ci"}
	uninstallOpts.Tag = p.Manifest.Dependencies["mysql"].Tag
	err := uninstallOpts.Validate([]string{"wordpress-mysql"}, p.Context)
	assert.NoError(p.T(), err, "validation of uninstall opts failed for dependent bundle")

	err = p.UninstallBundle(uninstallOpts)
	assert.NoError(p.T(), err, "uninstall failed for dependent bundle")

	// Uninstall the bundle
	uninstallOpts = porter.UninstallOptions{}
	uninstallOpts.CredentialIdentifiers = []string{"ci"}
	err = uninstallOpts.Validate([]string{}, p.Context)
	assert.NoError(p.T(), err, "validation of uninstall opts failed for dependent bundle")

	err = p.UninstallBundle(uninstallOpts)
	assert.NoError(p.T(), err, "uninstall failed for root bundle")
}

func upgradeWordpressBundle(p *porter.TestPorter, namespace string) {
	upgradeOpts := porter.UpgradeOptions{}
	upgradeOpts.CredentialIdentifiers = []string{"ci"}
	upgradeOpts.Params = []string{ // See https://github.com/deislabs/porter/issues/474
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
		// TODO: Remove this once we put back in dependency parameter defaults in #800
		"mysql#mysql-user=wordpress",
	}
	err := upgradeOpts.Validate([]string{}, p.Context)
	require.NoError(p.T(), err, "validation of upgrade opts for root bundle failed")

	err = p.UpgradeBundle(upgradeOpts)
	require.NoError(p.T(), err, "upgrade of root bundle failed")

	// Verify that the dependency claim is upgraded
	c, err := p.Claims.Read("wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")
	assert.Equal(p.T(), claim.ActionUpgrade, c.Result.Action, "the dependency wasn't recorded as being upgraded")
	assert.Equal(p.T(), claim.StatusSuccess, c.Result.Status, "the dependency wasn't recorded as being upgraded successfully")

	// Verify that the bundle claim is upgraded
	c, err = p.Claims.Read("wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	assert.Equal(p.T(), claim.ActionUpgrade, c.Result.Action, "the root bundle wasn't recorded as being upgraded")
	assert.Equal(p.T(), claim.StatusSuccess, c.Result.Status, "the root bundle wasn't recorded as being upgraded successfully")
}

func invokeWordpressBundle(p *porter.TestPorter, namespace string) {
	invokeOpts := porter.InvokeOptions{Action: "ping"}
	invokeOpts.CredentialIdentifiers = []string{"ci"}
	invokeOpts.Params = []string{ // See https://github.com/deislabs/porter/issues/474
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
		// TODO: Remove this once we put back in dependency parameter defaults in #800
		"mysql#mysql-user=wordpress",
	}
	err := invokeOpts.Validate([]string{}, p.Context)
	require.NoError(p.T(), err, "validation of invoke opts for root bundle failed")

	err = p.InvokeBundle(invokeOpts)
	require.NoError(p.T(), err, "invoke of root bundle failed")

	// Verify that the dependency claim is invoked
	c, err := p.Claims.Read("wordpress-mysql")
	require.NoError(p.T(), err, "could not fetch claim for the dependency")
	assert.Equal(p.T(), "ping", c.Result.Action, "the dependency wasn't recorded as being invoked")
	assert.Equal(p.T(), claim.StatusSuccess, c.Result.Status, "the dependency wasn't recorded as being invoked successfully")

	// Verify that the bundle claim is invoked
	c, err = p.Claims.Read("wordpress")
	require.NoError(p.T(), err, "could not fetch claim for the root bundle")
	assert.Equal(p.T(), "ping", c.Result.Action, "the root bundle wasn't recorded as being invoked")
	assert.Equal(p.T(), claim.StatusSuccess, c.Result.Status, "the root bundle wasn't recorded as being invoked successfully")
}

func uninstallWordpressBundle(p *porter.TestPorter, namespace string) {
	uninstallOptions := porter.UninstallOptions{}
	uninstallOptions.CredentialIdentifiers = []string{"ci"}
	uninstallOptions.Params = []string{ // See https://github.com/deislabs/porter/issues/474
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
		// TODO: Remove this once we put back in dependency parameter defaults in #800
		"mysql#mysql-user=wordpress",
	}
	err := uninstallOptions.Validate([]string{}, p.Context)
	require.NoError(p.T(), err, "validation of uninstall opts for root bundle failed")

	err = p.UninstallBundle(uninstallOptions)
	require.NoError(p.T(), err, "uninstall of root bundle failed")

	// Verify that the dependency claim is uninstalled
	_, err = p.Claims.Read("wordpress-mysql")
	assert.EqualError(p.T(), err, "Claim does not exist")

	// Verify that the bundle claim is uninstalled
	_, err = p.Claims.Read("wordpress")
	assert.EqualError(p.T(), err, "Claim does not exist")
}
