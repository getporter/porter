//go:build smoke

package smoke

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Test that telemetry data is being exported
func TestTelemetry(t *testing.T) {
	// I am always using require, so that we stop immediately upon an error
	// A long test is hard to debug when it fails in the middle and keeps going
	test, err := tester.NewTestWithConfig(t, "tests/testdata/config/config-with-telemetry.yaml")
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Enable telemetry
	err = shx.Copy(filepath.Join(test.RepoRoot, "tests/testdata/config/config-with-telemetry.yaml"), filepath.Join(test.PorterHomeDir, "config.yaml"))
	require.NoError(t, err, "error copying config file into PORTER_HOME")

	// Make a call that will call a plugin
	_, output, err := test.RunPorter("list")
	fmt.Println(output)
	require.NoError(t, err, "porter list failed")

	// Read the traces generated for that call
	tracesDir := filepath.Join(test.PorterHomeDir, "traces")
	traces, err := os.ReadDir(tracesDir)
	require.NoError(t, err, "error getting a list of the traces directory in PORTER_HOME")
	require.Len(t, traces, 2, "expected 2 trace files to be exported")

	// Validate we have trace data for porter (files are returned in descending order, which is why we know which to read first)
	porterTraceName := filepath.Join(tracesDir, traces[1].Name())
	porterTrace, err := ioutil.ReadFile(porterTraceName)
	require.NoError(t, err, "error reading porter's trace file %s", porterTraceName)
	tests.RequireOutputContains(t, string(porterTrace), `{"Key":"service.name","Value":{"Type":"STRING","Value":"porter"}}`, "no spans for porter were exported")

	// Validate we have trace data for porter
	pluginTraceName := filepath.Join(tracesDir, traces[0].Name())
	require.Contains(t, pluginTraceName, "storage.porter.mongodb", "expected the plugin trace to be for the mongodb plugin")
	pluginTrace, err := ioutil.ReadFile(pluginTraceName)
	require.NoError(t, err, "error reading the plugin's trace file %s", pluginTraceName)
	tests.RequireOutputContains(t, string(pluginTrace), `{"Key":"service.name","Value":{"Type":"STRING","Value":"storage.porter.mongodb"}}`, "no spans for the plugins were exported")
}
