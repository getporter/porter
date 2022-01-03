// +build smoke

package smoke

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Start up another docker registry to host the original bundle
// Publish a bundle to the temporary registry
// Copy the bundle to our integration test registry
// Stop the temporary registry
// Copy the bundle to another location, this will fail unless we are properly using the relocation map
func TestCopy(t *testing.T) {
	// Start a temp registry
	tempRegistryId, err := shx.OutputE("docker", "run", "-d", "-P", "registry:2")
	require.NoError(t, err, "Could not start a temporary registry")
	stopTempRegistry := func() error {
		return shx.RunE("docker", "rm", "-vf", tempRegistryId)
	}
	defer stopTempRegistry()

	// Get the porter that its running on
	tempRegistryPort, err := shx.OutputE("docker", "inspect", tempRegistryId, "--format", `{{ (index (index .NetworkSettings.Ports "5000/tcp") 0).HostPort }}`)
	require.NoError(t, err, "Could not get the published port of the temporary registry")

	test, err := tester.NewTest(t)
	defer test.Teardown()
	require.NoError(t, err, "test setup failed")

	// Build an interesting test bundle
	origRef := fmt.Sprintf("localhost:%s/mybuns:v0.1.1", tempRegistryPort)
	require.NoError(t, shx.Copy(filepath.Join(test.RepoRoot, "tests/testdata/mybuns"), ".", shx.CopyRecursive))
	os.Chdir("mybuns")
	test.RequirePorter("build")
	test.RequirePorter("publish", "--reference", origRef)

	// Copy the bundle to the integration test registry
	copiedRef := "localhost:5000/copy-mybuns:v0.1.1"
	test.RequirePorter("copy", "--source", origRef, "--destination", copiedRef)

	stopTempRegistry()

	// Copy the copied bundle to a new location. This will fail if we aren't using the relocation map.
	finalRef := "localhost:5000/copy-copy-mybuns:v0.1.1"
	test.RequirePorter("copy", "--source", copiedRef, "--destination", finalRef)

	// Get the original image from the relocation map
	inspectOutput, _, err := test.RunPorter("inspect", "-r", finalRef, "--output=json")
	var inspectRaw map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(inspectOutput), &inspectRaw))
	images := inspectRaw["invocationImages"].([]interface{})
	invocationImage := images[0].(map[string]interface{})
	require.Contains(t, invocationImage["originalImage"].(string), fmt.Sprintf("localhost:%s/mybuns-installer:v0.1.1", tempRegistryPort))
}
