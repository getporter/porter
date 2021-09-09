// +build smoke

package smoke

import (
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Test desired state workflows used by the porter operator
func TestDesiredState(t *testing.T) {
	// I am always using require, so that we stop immediately upon an error
	// A long test is hard to debug when it fails in the middle and keeps going
	test, err := tester.NewTest(t)
	defer test.Teardown()
	require.NoError(t, err, "test setup failed")

	test.PrepareTestBundle()
	require.NoError(t, shx.Copy("testdata/imported-buns.yaml", test.TestDir))
	os.Chdir(test.TestDir)

	// Try to import an installation with an invalid schema
	_, err = test.RunPorter("installation", "apply", filepath.Join(test.RepoRoot, "tests/testdata/installations/invalid-schema.yaml"))
	require.Error(t, err, "apply should have failed because the schema of the imported document is incorrect")
	require.Contains(t, err.Error(), "invalid installation")

	// Try to import a credential set with an invalid schema
	_, err = test.RunPorter("credentials", "apply", filepath.Join(test.RepoRoot, "tests/testdata/creds/invalid-schema.yaml"))
	require.Error(t, err, "apply should have failed because the schema of the imported document is incorrect")
	require.Contains(t, err.Error(), "invalid credential set")

	// Try to import a parameter set with an invalid schema
	_, err = test.RunPorter("parameters", "apply", filepath.Join(test.RepoRoot, "tests/testdata/params/invalid-schema.yaml"))
	require.Error(t, err, "apply should have failed because the schema of the imported document is incorrect")
	require.Contains(t, err.Error(), "invalid parameter set")

	// Import an installation where the namespace is empty in the file
	test.RequirePorter("installation", "apply", "imported-buns.yaml", "--namespace", "operator")
	test.RequireInstallationExists("operator", "imported-buns")
}
