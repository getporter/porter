// +build smoke

package smoke

import (
	"os"
	"testing"

	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Test desired state workflows used by the porter operator
func TestDesiredState(t *testing.T) {
	// I am always using require, so that we stop immediately upon an error
	// A long test is hard to debug when it fails in the middle and keeps going
	test, err := NewTest(t)
	defer test.Teardown()
	require.NoError(t, err, "test setup failed")

	test.PrepareTestBundle()
	require.NoError(t, shx.Copy("testdata/imported-buns.yaml", test.TestDir))
	os.Chdir(test.TestDir)

	// Import an installation where the namespace is empty in the file
	test.RequirePorter("installation", "apply", "imported-buns.yaml", "--namespace", "operator")
	test.RequireInstallationExists("operator", "imported-buns")
}
