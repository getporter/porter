//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

func TestCopy_UsesRelocationMap(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Start a temporary registry, that uses plain http (no TLS)
	reg1 := test.StartTestRegistry(tester.TestRegistryOptions{UseTLS: false})

	// Publish the bundle to the insecure registry
	origRef := fmt.Sprintf("%s/orig-mydb:v0.1.1", reg1)
	test.MakeTestBundle(testdata.MyDb, origRef)

	ociRef, err := cnab.ParseOCIReference(origRef)
	require.NoError(t, err)
	out, _ := test.RequirePorter("install", "--force", "-r", origRef)
	require.Contains(t, out, ociRef.Repository())

	// Start a temporary (insecure) registry on a random port, with a self-signed certificate
	reg2 := test.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})

	// Copy the bundle to the integration test registry, using --insecure-registry
	// because the destination uses a self-signed certificate
	copiedRef := fmt.Sprintf("%s/copy-mydb:v0.1.1", reg2)
	test.RequirePorter("copy", "--source", origRef, "--destination", copiedRef, "--insecure-registry")

	reg1.Close()

	// Copy the copied bundle to a new location. This will fail if we aren't using the relocation map.
	finalRef := fmt.Sprintf("%s/copy-copy-mydb:v0.1.1", reg2)
	test.RequirePorter("copy", "--source", copiedRef, "--destination", finalRef, "--insecure-registry")

	finalOCIRef, err := cnab.ParseOCIReference(finalRef)
	require.NoError(t, err)
	finalOut, _ := test.RequirePorter("install", "--force", "-r", finalRef, "--insecure-registry")
	require.Contains(t, finalOut, finalOCIRef.Repository())

	// Get the original image from the relocation map
	inspectOutput, _, err := test.RunPorter("inspect", "-r", finalRef, "--output=json", "--insecure-registry")
	require.NoError(t, err, "porter inspect failed")
	var inspectRaw map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(inspectOutput), &inspectRaw))
	images := inspectRaw["invocationImages"].([]interface{})
	invocationImage := images[0].(map[string]interface{})
	require.Contains(t, invocationImage["originalImage"].(string), fmt.Sprintf("%s/orig-mydb", reg1))
}
