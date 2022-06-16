//go:build smoke
// +build smoke

package smoke

import (
	"fmt"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Start up another docker registry to host the original bundle
// Publish a bundle to the temporary registry
// Archive the bundle from the temporary registry
// Stop the temporary registry
// Publish the bundle from the archive file
func TestArchive(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Start a temp registry
	tempRegistryId, err := shx.OutputE("docker", "run", "-d", "-P", "registry:2")
	require.NoError(t, err, "Could not start a temporary registry")
	stopTempRegistry := func() error {
		return shx.RunE("docker", "rm", "-vf", tempRegistryId)
	}
	defer stopTempRegistry()

	// Get the port that it is running on
	tempRegistryPort, err := shx.OutputE("docker", "inspect", tempRegistryId, "--format", `{{ (index (index .NetworkSettings.Ports "5000/tcp") 0).HostPort }}`)
	require.NoError(t, err, "Could not get the published port of the temporary registry")

	// Publish the bundle to one location
	origRef := fmt.Sprintf("localhost:%s/orig-mydb:v0.1.1", tempRegistryPort)
	test.MakeTestBundle(testdata.MyDb, origRef)

	archiveFilePath := filepath.Join(test.TestDir, "archive-test.tgz")
	test.RequirePorter("archive", archiveFilePath, "--reference", origRef)
	stopTempRegistry()

	newRef := "localhost:5000/copy-new-mydb:v0.1.1"
	test.RequirePorter("publish", "--archive", archiveFilePath, "--reference", newRef)
}
