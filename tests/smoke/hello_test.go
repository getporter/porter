// +build smoke

package smoke

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelloBundle(t *testing.T) {
	test, err := NewTest(t)
	defer test.Teardown()
	require.NoError(t, err, "test setup failed")

	bundleDir := filepath.Join(test.TestDir, "hello")
	os.MkdirAll(bundleDir, 0700)
	os.Chdir(bundleDir)

	test.RequirePorter("create")
	test.RequirePorter("build")

	ref := "localhost:5000/porter-hello:v0.1.1"
	test.RequirePorter("publish", "--reference", ref)

	os.Chdir(test.TestDir)

	test.RequirePorter("install", "hello", "--reference", ref, "--namespace=dev")

	// Should not see the installation in the global namespace
	output, err := test.Porter("list", "--output=json").Output()
	require.NoError(t, err)
	var installations []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &installations))
	assert.Empty(t, installations, "expected no global installations to exist")

	// Should see the installation in the dev namespace
	output, err = test.Porter("list", "--namespace=dev", "--output=json").Output()
	require.NoError(t, err)
	installations = []map[string]interface{}{}
	require.NoError(t, json.Unmarshal([]byte(output), &installations))
	require.Len(t, installations, 1, "expected one installation to be returned by porter list")
	assert.Equal(t, "hello", installations[0]["name"], "expected the hello installation to be output by porter list")
	assert.Equal(t, "dev", installations[0]["namespace"], "expected the installation to be in the dev namespace")

	// The installation should be successful
	status := installations[0]["status"].(map[string]interface{})
	assert.Equal(t, true, status["installationCompleted"], "expected the installation to complete successfully")

	test.RequirePorter("install", "hello", "--reference", ref, "--namespace=test")

	// Should see the other installation in the test namespace
	output, err = test.Porter("list", "--namespace=test", "--output=json").Output()
	require.NoError(t, err)
	installations = []map[string]interface{}{}
	require.NoError(t, json.Unmarshal([]byte(output), &installations))
	require.Len(t, installations, 1, "expected one installation to be returned by porter list")
	assert.Equal(t, "hello", installations[0]["name"], "expected the hello installation to be output by porter list")
	assert.Equal(t, "test", installations[0]["namespace"], "expected the installation to be in the test namespace")

	// Validate that we can't accidentally overwrite an installation
	var outputE bytes.Buffer
	_, _, err = test.Porter("install", "hello", "--reference", ref, "--namespace=dev").Stderr(&outputE).Exec()
	require.Error(t, err, "porter should have prevented overwriting an installation")
	assert.Contains(t, outputE.String(), "The installation has already been successfully installed")

	test.RequirePorter("installation", "show", "hello", "--namespace=dev")

	test.RequirePorter("upgrade", "hello", "--namespace=dev")
	test.RequirePorter("installation", "show", "hello", "--namespace=dev")

	test.RequirePorter("uninstall", "hello", "--namespace=dev")
	test.RequirePorter("installation", "show", "hello", "--namespace=dev")
	test.RequirePorter("installation", "delete", "hello", "--namespace=dev")

	outputE.Truncate(0)
	_, _, err = test.Porter("installation", "show", "hello", "--namespace=dev").Stderr(&outputE).Exec()
	require.Error(t, err, "show should fail because the installation was deleted")
	assert.Contains(t, outputE.String(), "not found")
}
