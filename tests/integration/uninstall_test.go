//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

func TestUninstall_DeleteInstallation(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")
	test.PrepareTestBundle()

	// Check that we can't uninstall a bundle that hasn't been installed already
	_, _, err = test.RunPorter("uninstall", testdata.MyBuns, "-c=mybuns")
	test.RequireNotFoundReturned(err)

	// Install bundle
	test.RequirePorter("install", testdata.MyBuns, "-r", testdata.MyBunsRef, "-c=mybuns", "--param", "password=supersecret")

	// Uninstall the bundle
	test.RequirePorter("uninstall", testdata.MyBuns, "-c=mybuns", "--param", "password=supersecret")

	// Check that the record remains
	test.RequireInstallationExists(test.CurrentNamespace(), testdata.MyBuns)

	// Uninstall and delete
	test.RequirePorter("uninstall", testdata.MyBuns, "-c=mybuns", "--delete", "--param", "password=supersecret")

	// The record should be gone
	test.RequireInstallationNotFound(test.CurrentNamespace(), testdata.MyBuns)

	// Re-Install the bundle
	test.RequirePorter("install", testdata.MyBuns, "-r", testdata.MyBunsRef, "-c=mybuns", "--param", "password=supersecret")

	// Uninstall the bundle, attempt to delete it, but have the uninstall fail
	_, _, err = test.RunPorter("uninstall", testdata.MyBuns, "-c=mybuns", "--param", "chaos_monkey=true", "--delete", "--param", "password=supersecret")
	tests.RequireErrorContains(t, err, "it is unsafe to delete an installation when the last action wasn't a successful uninstall")

	// Check that the record remains
	test.RequireInstallationExists(test.CurrentNamespace(), testdata.MyBuns)

	// Uninstall the bundle, even though uninstall is failing, and force delete it
	test.RequirePorter("uninstall", testdata.MyBuns, "-c=mybuns", "--force-delete", "--param", "password=supersecret")

	// The record should be gone
	test.RequireInstallationNotFound(test.CurrentNamespace(), testdata.MyBuns)
}
