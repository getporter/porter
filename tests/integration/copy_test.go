//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

func TestCopy_UsesRelocationMap(t *testing.T) {
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

	// Copy the bundle to the integration test registry
	copiedRef := "localhost:5000/copy-mydb:v0.1.1"
	test.RequirePorter("copy", "--source", origRef, "--destination", copiedRef)

	stopTempRegistry()

	// Copy the copied bundle to a new location. This will fail if we aren't using the relocation map.
	finalRef := "localhost:5000/copy-copy-mydb:v0.1.1"
	test.RequirePorter("copy", "--source", copiedRef, "--destination", finalRef)

	// Get the original image from the relocation map
	inspectOutput, _, err := test.RunPorter("inspect", "-r", finalRef, "--output=json")
	var inspectRaw map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(inspectOutput), &inspectRaw))
	images := inspectRaw["invocationImages"].([]interface{})
	invocationImage := images[0].(map[string]interface{})
	require.Contains(t, invocationImage["originalImage"].(string), fmt.Sprintf("localhost:%s/orig-mydb", tempRegistryPort))
}
