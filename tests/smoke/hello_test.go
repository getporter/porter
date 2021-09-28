// +build smoke

package smoke

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

func TestHelloBundle(t *testing.T) {
	test, err := NewTest(t)
	defer test.Teardown()
	require.NoError(t, err, "test setup failed")

	// Build an interesting test bundle
	ref := "localhost:5000/mybuns:v0.1.1"
	shx.Copy("../testdata/mybuns", ".", shx.CopyRecursive)
	os.Chdir("mybuns")
	test.RequirePorter("build")
	test.RequirePorter("publish", "--reference", ref)

	// Do not run these commands in a bundle directory
	os.Chdir(test.TestDir)

	test.RequirePorter("install", "--reference", ref)
	test.RequirePorter("installation", "show", "mybuns")
	test.RequirePorter("installation", "logs", "show", "-i=mybuns")

	test.RequirePorter("upgrade", "mybuns")
	test.RequirePorter("installation", "show", "mybuns")

	// Check that we can disable logs from persisting
	test.RequirePorter("uninstall", "mybuns", "--no-logs")
	test.RequirePorter("installation", "show", "mybuns")
	_, _, err = test.RunPorter("installation", "logs", "show", "-i=mybuns")
	require.Error(t, err, "expected log retrieval to fail")
	require.Contains(t, err.Error(), "no logs found")

	// Verify file permissions on PORTER_HOME
	test.RequireFileMode(filepath.Join(test.PorterHomeDir, "claims", "*"), 0600)
	test.RequireFileMode(filepath.Join(test.PorterHomeDir, "results", "*"), 0600)
	test.RequireFileMode(filepath.Join(test.PorterHomeDir, "outputs", "*"), 0600)
}
