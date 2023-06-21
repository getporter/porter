//go:build smoke

package smoke

import (
	"testing"

	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Test general flows in porter
func TestHelloBundle(t *testing.T) {
	// I am always using require, so that we stop immediately upon an error
	// A long test is hard to debug when it fails in the middle and keeps going
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	test.PrepareTestBundle()
	require.NoError(t, shx.Copy("testdata/buncfg.json", test.TestDir))
	require.NoError(t, shx.Copy("testdata/plugins.yaml", test.TestDir))
	test.Chdir(test.TestDir)

	// Verify plugins installation
	// This also does a quick regression test to validate that we can change the verbosity to a different value than what is in the config file and have it respected
	_, output := test.RequirePorter("plugins", "install", "-f", "plugins.yaml", "--verbosity=info")
	require.Contains(t, output, "installed azure plugin", "expected to see plugin successfully installed")
	require.NotContainsf(t, output, "Downloading https://cdn.porter.sh/plugins/atom.xml", "Debug information should not have been printed because verbosity is set to info")

	// Run a stateless action before we install and make sure nothing is persisted
	_, output = test.RequirePorter("invoke", testdata.MyBuns, "--action=dry-run", "--reference", testdata.MyBunsRef, "-c=mybuns")
	t.Log(output)
	test.RequireInstallationNotFound(test.CurrentNamespace(), testdata.MyBuns)

	// Install the bundle and verify the correct output is printed
	_, output = test.RequirePorter("install", testdata.MyBuns, "--reference", testdata.MyBunsRef, "--label", "test=true", "-p=mybuns", "-c=mybuns", "--param", "password=supersecret")
	require.Contains(t, output, "Hello, *******", "expected to see output printed from the bundle itself")
	// Make sure that when the mixin uses span.Debug to print the command it is running, that it's being printed
	require.Contains(t, output, "/cnab/app ./helpers.sh install", "expected to see output printed from the porter runtime libraries")

	// Should not see the mybuns installation in the global namespace
	test.RequireInstallationNotFound("", testdata.MyBuns)

	// Should see the installation in the current namespace, it should be successful
	installation := test.RequireInstallationExists(test.CurrentNamespace(), testdata.MyBuns)
	require.Equal(t, "succeeded", installation.Status.ResultStatus)

	// Logs should have been persisted for the run
	test.RequirePorter("installation", "logs", "show", "-i=mybuns")

	// Run a no-op action to check the status and check that the run was persisted
	// Also checks that we are processing file parameters properly, when templated and read from the filesystem
	_, output = test.RequirePorter("invoke", testdata.MyBuns, "--action=status", "-c=mybuns", "--param", "cfg=./buncfg.json")
	require.Contains(t, output, `{"color": "blue"}`, "templated file parameter was not decoded properly")
	require.Contains(t, output, `is a unicorn`, "state file parameter was not decoded properly")

	// Check that the last action is still install, a noop action shouldn't update the installation status
	installation = test.RequireInstallationExists(test.CurrentNamespace(), testdata.MyBuns)
	require.Equal(t, "install", installation.Status.Action) // Install should be the last modifying action
	// TODO(carolynvs): check that status shows up as a run

	// Install in the test namespace, and do not persist the logs
	test.RequirePorter("install", testdata.MyBuns, "--reference", testdata.MyBunsRef, "--namespace=test", "-c=mybuns", "-p=mybuns", "--no-logs", "--param", "password=supersecret")
	_, _, err = test.RunPorter("installation", "logs", "show", "--namespace=test", "-i=mybuns")
	require.Error(t, err, "expected log retrieval to fail")
	require.Contains(t, err.Error(), "no logs found")
	displayInstallation, err := test.ShowInstallation("test", testdata.MyBuns)
	require.NoError(t, err)
	require.Len(t, displayInstallation.ParameterSets, 1)

	// Let's try out list filtering!
	// Search by namespace
	installations, err := test.ListInstallations(false, "test", "", nil)
	require.NoError(t, err)
	require.Len(t, installations, 2, "expected two installations in the test namespace: mybuns-db and mybuns")
	require.Equal(t, "mybuns-db", installations[0].Name)
	require.Equal(t, "mybuns", installations[1].Name)

	// Search by name
	installations, err = test.ListInstallations(true, "", testdata.MyBuns, nil)
	require.NoError(t, err)
	require.Len(t, installations, 4, "expected four installations with mybuns in the name")

	// Search by label
	installations, err = test.ListInstallations(true, "", "", []string{"test=true"})
	require.NoError(t, err)
	require.Len(t, installations, 1, "expected one installations labeled with test=true")

	// Validate that we can't accidentally overwrite an installation
	_, _, err = test.RunPorter("install", testdata.MyBuns, "--reference", testdata.MyBunsRef, "--namespace=test", "-c=mybuns", "--param", "password=supersecret")
	tests.RequireErrorContains(t, err, "The installation has already been successfully installed")

	// We should be able to repeat install with --force
	test.RequirePorter("install", testdata.MyBuns, "--reference", testdata.MyBunsRef, "--namespace=test", "-c=mybuns", "--force", "--param", "password=supersecret")

	// Upgrade our installation, passing the same cred/param set that is already specified, it shouldn't create duplicates
	// We are also overridding a different parameter, the old value for password should be remembered even though we didn't explicitly set it
	test.RequirePorter("upgrade", testdata.MyBuns, "--namespace", test.CurrentNamespace(), "-p=mybuns", "-c=mybuns", "--param", "log_level=1")
	// no duplicate in credential set or parameter set on the installation
	// record
	displayInstallation, err = test.ShowInstallation(test.CurrentNamespace(), testdata.MyBuns)
	require.NoError(t, err)
	require.Len(t, displayInstallation.ParameterSets, 1)
	require.Len(t, displayInstallation.CredentialSets, 1)
	param, ok := displayInstallation.ResolvedParameters.Get("password")
	require.True(t, ok, "Could not find resolved parameter 'password'")
	require.Equal(t, "supersecret", param.Value, "Expected that the previous override provided during install should be remembered and reused during upgrade")
	param, ok = displayInstallation.ResolvedParameters.Get("log_level")
	require.True(t, ok, "Could not find resolved parameter 'log_level'")
	require.Equal(t, float64(1), param.Value, "Expected the new override for log_level to be used")

	// Uninstall and remove the installation
	test.RequirePorter("uninstall", testdata.MyBuns, "--namespace", test.CurrentNamespace(), "-c=mybuns")
	displayInstallations, err := test.ListInstallations(false, test.CurrentNamespace(), testdata.MyBuns, nil)
	require.NoError(t, err, "List installations failed")
	require.Len(t, displayInstallations, 2, "expected the installations to still be returned by porter list even though it's uninstalled")
	require.NotEmpty(t, displayInstallations[0].Status.Uninstalled, "expected the installations to be flagged as uninstalled")
	require.NotEmpty(t, displayInstallations[1].Status.Uninstalled, "expected the installations to be flagged as uninstalled")

	test.RequirePorter("installation", "delete", testdata.MyBuns, "--namespace", test.CurrentNamespace())
	test.RequireInstallationNotFound(test.CurrentNamespace(), testdata.MyBuns)

	// Let's test some negatives!

	// Cannot perform a modifying or stateful action without an installation
	_, _, err = test.RunPorter("upgrade", "missing", "--reference", testdata.MyBunsRef)
	test.RequireNotFoundReturned(err)

	_, _, err = test.RunPorter("uninstall", "missing", "--reference", testdata.MyBunsRef)
	test.RequireNotFoundReturned(err)

	_, _, err = test.RunPorter("invoke", "--action=boom", "missing", "--reference", testdata.MyBunsRef)
	test.RequireNotFoundReturned(err)

	_, _, err = test.RunPorter("invoke", "--action=status", "missing", "--reference", testdata.MyBunsRef)
	test.RequireNotFoundReturned(err)

	// Test that outputs are collected when a bundle fails
	_, _, err = test.RunPorter("install", "fail-with-outputs", "--reference", testdata.MyBunsRef, "-c=mybuns", "-p=mybuns", "--param", "chaos_monkey=true")
	require.Error(t, err, "the chaos monkey should have failed the installation")
	myLogs, _ := test.RequirePorter("installation", "outputs", "show", "mylogs", "-i=fail-with-outputs")
	require.Contains(t, myLogs, "Hello, porterci")

	myLogsListed, _ := test.RequirePorter("installation", "outputs", "list", "-i=fail-with-outputs")
	require.Contains(t, myLogsListed, "Hello, porterci")
}
