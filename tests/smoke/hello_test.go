// +build smoke

package smoke

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

func TestHelloBundle(t *testing.T) {
	// I am always using require, so that we stop immediately upon an error
	// A long test is hard to debug when it fails in the middle and keeps going

	test, err := NewTest(t)
	defer test.Teardown()
	require.NoError(t, err, "test setup failed")

	// Build an interesting test bundle
	ref := "localhost:5000/mybuns:v0.1.1"
	shx.Copy("../testdata/mybuns", test.TestDir, shx.CopyRecursive)
	os.Chdir(filepath.Join(test.TestDir, "mybuns"))
	test.RequirePorter("build")
	test.RequirePorter("publish", "--reference", ref)

	// Do not run these commands in a bundle directory
	os.Chdir(test.TestDir)

	test.RequirePorter("install", "mybuns", "--reference", ref, "--namespace=dev", "--label", "test=true")

	// Should not see the installation in the global namespace
	output, err := test.Porter("list", "--output=json").Output()
	require.NoError(t, err)
	var installations []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &installations))
	require.Empty(t, installations, "expected no global installations to exist")

	// Should see the installation in the dev namespace
	output, err = test.Porter("list", "--namespace=dev", "--output=json").Output()
	require.NoError(t, err)
	installations = []map[string]interface{}{}
	require.NoError(t, json.Unmarshal([]byte(output), &installations))
	require.Len(t, installations, 1, "expected one installation to be returned by porter list")
	require.Equal(t, "mybuns", installations[0]["name"], "expected the mybuns installation to be output by porter list")
	require.Equal(t, "dev", installations[0]["namespace"], "expected the installation to be in the dev namespace")

	// The installation should be successful
	status := installations[0]["status"].(map[string]interface{})
	require.Equal(t, true, status["installationCompleted"], "expected the installation to complete successfully")

	test.RequirePorter("install", "mybuns", "--reference", ref, "--namespace=test")

	// Should see the other installation in the test namespace
	output, err = test.Porter("list", "--namespace=test", "--output=json").Output()
	require.NoError(t, err)
	installations = []map[string]interface{}{}
	require.NoError(t, json.Unmarshal([]byte(output), &installations))
	require.Len(t, installations, 1, "expected one installation to be returned by porter list")
	require.Equal(t, "mybuns", installations[0]["name"], "expected the mybuns installation to be output by porter list")
	require.Equal(t, "test", installations[0]["namespace"], "expected the installation to be in the test namespace")

	// Search by name
	output, err = test.Porter("list", "mybuns", "--namespace=*", "--output=json").Output()
	require.NoError(t, err)
	installations = []map[string]interface{}{}
	require.NoError(t, json.Unmarshal([]byte(output), &installations))
	require.Len(t, installations, 2, "expected two installations named mybuns")

	// Search by label
	output, err = test.Porter("list", "--label", "test=true", "--namespace=*", "--output=json").Output()
	require.NoError(t, err)
	installations = []map[string]interface{}{}
	require.NoError(t, json.Unmarshal([]byte(output), &installations))
	require.Len(t, installations, 1, "expected one installations labeled with test=true")

	// Validate that we can't accidentally overwrite an installation
	var outputE bytes.Buffer
	_, _, err = test.Porter("install", "mybuns", "--reference", ref, "--namespace=dev").Stderr(&outputE).Exec()
	require.Error(t, err, "porter should have prevented overwriting an installation")
	require.Contains(t, outputE.String(), "The installation has already been successfully installed")

	test.RequirePorter("installation", "show", "mybuns", "--namespace=dev")

	test.RequirePorter("upgrade", "mybuns", "--namespace=dev")
	test.RequirePorter("installation", "show", "mybuns", "--namespace=dev")

	test.RequirePorter("uninstall", "mybuns", "--namespace=dev")
	test.RequirePorter("installation", "show", "mybuns", "--namespace=dev")
	test.RequirePorter("installation", "delete", "mybuns", "--namespace=dev")

	outputE.Truncate(0)
	_, _, err = test.Porter("installation", "show", "mybuns", "--namespace=dev").Stderr(&outputE).Exec()
	require.Error(t, err, "show should fail because the installation was deleted")
	require.Contains(t, outputE.String(), "not found")
}
