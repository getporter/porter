// +build integration

package tests

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		//A=97 and Z = 97+25
		bytes[i] = byte(97 + rand.Intn(25))
	}
	return string(bytes)
}

func publishMySQLBundle(p *porter.TestPorter) {
	mysqlBundlePath := filepath.Join(p.TestDir, "../build/testdata/bundles/mysql/porter.yaml")
	p.CopyFile(mysqlBundlePath, "porter.yaml")

	publishOpts := porter.PublishOptions{}
	err := publishOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for dependent bundle failed")

	err = p.Publish(publishOpts)
	require.NoError(p.T(), err, "publish of dependent bundle failed")
}

func installWordpressBundle(p *porter.TestPorter) (namespace string) {
	// Publish the mysql bundle that we depend upon
	publishMySQLBundle(p)

	// Install the bundle that has dependencies
	p.CopyFile(filepath.Join(p.TestDir, "../build/testdata/bundles/wordpress/porter.yaml"), "porter.yaml")

	namespace = randomString(10)
	installOpts := porter.InstallOptions{}
	installOpts.CredentialIdentifiers = []string{"ci"}
	installOpts.Params = []string{
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
	}
	err := installOpts.Validate([]string{}, p.Context)
	require.NoError(p.T(), err, "validation of install opts for root bundle failed")

	err = p.InstallBundle(installOpts)
	require.NoError(p.T(), err, "install of root bundle failed")

	return namespace
}

func TestDependencies_Install(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	installWordpressBundle(p)

	// Verify that the dependency claim is present
	c, err := p.CNAB.FetchClaim("wordpress-mysql")
	require.NoError(t, err, "could not fetch claim for the dependency")
	assert.Equal(t, claim.StatusSuccess, c.Result.Status, "the dependency wasn't recorded as being installed successfully")

	// Verify that the bundle claim is present
	c, err = p.CNAB.FetchClaim("wordpress")
	require.NoError(t, err, "could not fetch claim for the root bundle")
	assert.Equal(t, claim.StatusSuccess, c.Result.Status, "the root bundle wasn't recorded as being installed successfully")

	// Uninstall the dependency
	uninstallOpts := porter.UninstallOptions{}
	uninstallOpts.CredentialIdentifiers = []string{"ci"}
	uninstallOpts.Tag = p.Manifest.Dependencies["mysql"].Tag
	err = uninstallOpts.Validate([]string{"wordpress-mysql"}, p.Context)
	require.NoError(t, err, "validation of uninstall opts failed for dependent bundle")

	err = p.UninstallBundle(uninstallOpts)
	require.NoError(t, err, "uninstall failed for dependent bundle")

	// Uninstall the bundle
	uninstallOpts = porter.UninstallOptions{}
	uninstallOpts.CredentialIdentifiers = []string{"ci"}
	err = uninstallOpts.Validate([]string{}, p.Context)
	require.NoError(t, err, "validation of uninstall opts failed for dependent bundle")

	err = p.UninstallBundle(uninstallOpts)
	require.NoError(t, err, "uninstall failed for root bundle")
}

func TestDependencies_Upgrade(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	namespace := installWordpressBundle(p)

	upgradeOpts := porter.UpgradeOptions{}
	upgradeOpts.CredentialIdentifiers = []string{"ci"}
	upgradeOpts.Params = []string{ // See https://github.com/deislabs/porter/issues/474
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
	}
	err := upgradeOpts.Validate([]string{}, p.Context)
	require.NoError(t, err, "validation of upgrade opts for root bundle failed")

	err = p.UpgradeBundle(upgradeOpts)
	require.NoError(t, err, "upgrade of root bundle failed")

	// Verify that the dependency claim is upgraded
	c, err := p.CNAB.FetchClaim("wordpress-mysql")
	require.NoError(t, err, "could not fetch claim for the dependency")
	assert.Equal(t, claim.ActionUpgrade, c.Result.Action, "the dependency wasn't recorded as being upgraded")
	assert.Equal(t, claim.StatusSuccess, c.Result.Status, "the dependency wasn't recorded as being upgraded successfully")

	// Verify that the bundle claim is upgraded
	c, err = p.CNAB.FetchClaim("wordpress")
	require.NoError(t, err, "could not fetch claim for the root bundle")
	assert.Equal(t, claim.ActionUpgrade, c.Result.Action, "the root bundle wasn't recorded as being upgraded")
	assert.Equal(t, claim.StatusSuccess, c.Result.Status, "the root bundle wasn't recorded as being upgraded successfully")

	// Uninstall the dependency
	uninstallOpts := porter.UninstallOptions{}
	uninstallOpts.CredentialIdentifiers = []string{"ci"}
	uninstallOpts.Tag = p.Manifest.Dependencies["mysql"].Tag
	err = uninstallOpts.Validate([]string{"wordpress-mysql"}, p.Context)
	require.NoError(t, err, "validation of uninstall opts failed for dependent bundle")

	err = p.UninstallBundle(uninstallOpts)
	require.NoError(t, err, "uninstall failed for dependent bundle")

	// Uninstall the bundle
	uninstallOpts = porter.UninstallOptions{}
	uninstallOpts.CredentialIdentifiers = []string{"ci"}
	err = uninstallOpts.Validate([]string{}, p.Context)
	require.NoError(t, err, "validation of uninstall opts failed for dependent bundle")

	err = p.UninstallBundle(uninstallOpts)
	require.NoError(t, err, "uninstall failed for root bundle")
}

func TestDependencies_Uninstall(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	namespace := installWordpressBundle(p)

	uninstallOptions := porter.UninstallOptions{}
	uninstallOptions.CredentialIdentifiers = []string{"ci"}
	uninstallOptions.Params = []string{ // See https://github.com/deislabs/porter/issues/474
		"wordpress-password=mypassword",
		"namespace=" + namespace,
		"wordpress-name=porter-ci-wordpress-" + namespace,
		"mysql#mysql-name=porter-ci-mysql-" + namespace,
	}
	err := uninstallOptions.Validate([]string{}, p.Context)
	require.NoError(t, err, "validation of uninstall opts for root bundle failed")

	err = p.UninstallBundle(uninstallOptions)
	require.NoError(t, err, "uninstall of root bundle failed")

	// Verify that the dependency claim is uninstalled
	_, err = p.CNAB.FetchClaim("wordpress-mysql")
	assert.EqualError(t, err, "could not retrieve claim wordpress-mysql: Claim does not exist")

	// Verify that the bundle claim is uninstalled
	_, err = p.CNAB.FetchClaim("wordpress")
	assert.EqualError(t, err, "could not retrieve claim wordpress: Claim does not exist")
}
