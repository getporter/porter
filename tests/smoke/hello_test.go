// +build smoke

package smoke

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Test general flows in porter
func TestHelloBundle(t *testing.T) {
	// I am always using require, so that we stop immediately upon an error
	// A long test is hard to debug when it fails in the middle and keeps going
	test, err := NewTest(t)
	defer test.Teardown()
	require.NoError(t, err, "test setup failed")

	test.PrepareTestBundle()
	os.Chdir(test.TestDir)

	test.RequirePorter("install", "hello", "--reference", "getporter/porter-hello:v0.1.1", "--namespace=")

	// Import a parameter and credential set for the bundle
	shx.Copy(filepath.Join(test.RepoRoot, "tests/testdata/params/mybuns.yaml"), "./myparams.yaml")
	shx.Copy(filepath.Join(test.RepoRoot, "tests/testdata/creds/mybuns.yaml"), "./mycreds.yaml")
	test.RequirePorter("parameters", "apply", "myparams.yaml", "--namespace=")
	test.RequirePorter("credentials", "apply", "mycreds.yaml", "--namespace=")

	// Run a stateless action before we install and make sure nothing is persisted
	var outputE bytes.Buffer
	test.RequirePorter("invoke", "mybuns", "--action=dry-run", "--reference", myBunsRef, "-c=mybuns")
	test.RequireInstallationNotFound("dev", "mybuns")

	// Install the bundle and verify the correct output is printed
	output, err := test.Porter("install", "mybuns", "--reference", myBunsRef, "--label", "test=true", "-p=mybuns", "-c=mybuns").OutputV()
	require.NoError(t, err)
	require.Contains(t, output, "Hello, *******")

	// Should not see the mybuns installation in the global namespace
	test.RequireInstallationNotFound("", "mybuns")

	// Should see the installation in the dev namespace, it should be successful
	installation := test.RequireInstallationExists("dev", "mybuns")
	require.Equal(t, "succeeded", installation.Status.ResultStatus)

	// Run a no-op action to check the status and check that the run was persisted
	test.RequirePorter("invoke", "mybuns", "--action=status", "-c=mybuns")
	installation = test.RequireInstallationExists("dev", "mybuns")
	require.Equal(t, "install", installation.Status.Action) // Install should be the last modifying action
	// TODO(carolynvs): check that status shows up as a run

	// Install in the test namespace
	test.RequirePorter("install", "mybuns", "--reference", myBunsRef, "--namespace=test", "-c=mybuns")

	// Let's try out list filtering!
	// Search by namespace
	installations, err := test.ListInstallations(false, "test", "", nil)
	require.NoError(t, err)
	require.Len(t, installations, 1, "expected one installation in the test namespace")

	// Search by name
	installations, err = test.ListInstallations(true, "", "mybuns", nil)
	require.NoError(t, err)
	require.Len(t, installations, 2, "expected two installations named mybuns")

	// Search by label
	installations, err = test.ListInstallations(true, "", "", []string{"test=true"})
	require.NoError(t, err)
	require.Len(t, installations, 1, "expected one installations labeled with test=true")

	// Validate that we can't accidentally overwrite an installation
	outputE.Truncate(0)
	_, _, err = test.Porter("install", "mybuns", "--reference", myBunsRef, "--namespace=dev", "-c=mybuns").Stderr(&outputE).Exec()
	require.Error(t, err, "porter should have prevented overwriting an installation")
	require.Contains(t, outputE.String(), "The installation has already been successfully installed")

	// Upgrade our installation
	test.RequirePorter("upgrade", "mybuns", "--namespace=dev", "-c=mybuns")

	// Uninstall and remove the installation
	test.RequirePorter("uninstall", "mybuns", "--namespace=dev", "-c=mybuns")
	test.RequirePorter("installation", "delete", "mybuns", "--namespace=dev")
	test.RequireInstallationNotFound("dev", "mybuns")

	// Let's test some negatives!

	// Cannot perform a modifying or stateful action without an installation
	_, err = test.RunPorter("upgrade", "missing", "--reference", myBunsRef)
	test.RequireNotFoundReturned(err)

	_, err = test.RunPorter("uninstall", "missing", "--reference", myBunsRef)
	test.RequireNotFoundReturned(err)

	_, err = test.RunPorter("invoke", "--action=boom", "missing", "--reference", myBunsRef)
	test.RequireNotFoundReturned(err)

	_, err = test.RunPorter("invoke", "--action=status", "missing", "--reference", myBunsRef)
	test.RequireNotFoundReturned(err)
}
