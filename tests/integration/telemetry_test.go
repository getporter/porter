//go:build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
	"github.com/uwu-tools/magex/shx"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// Test that trace data is sent to the configured telemetry endpoint, both from porter and the plugins
func TestTelemetry_TracesExported(t *testing.T) {
	testcases := []struct {
		name     string
		protocol string
	}{
		{name: "grpc", protocol: tester.OTLPProtocolGRPC},
		{name: "http", protocol: tester.OTLPProtocolHTTP},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			test, err := tester.NewTest(t)
			defer test.Close()
			require.NoError(t, err, "test setup failed")

			receiver := test.StartTestOTLPReceiver(tc.protocol)

			// Make a call that will call a plugin
			test.RequirePorter("list")

			// Validate we have trace data for porter
			porterSpans := receiver.RequireSpans("porter")
			porterTraces := make(map[string]struct{}, len(porterSpans))
			for _, span := range porterSpans {
				porterTraces[span.TraceID()] = struct{}{}
			}

			// Validate we have trace data for the plugin, and that the calls made to it are part of the trace started by porter.
			// The plugin also exports spans in a separate trace for its own startup, e.g. loading its configuration.
			var linkedSpans int
			for _, span := range receiver.RequireSpans("storage.porter.mongodb") {
				if _, ok := porterTraces[span.TraceID()]; ok {
					linkedSpans++
				}
			}
			require.NotZero(t, linkedSpans, "expected spans from the plugin to be in the same trace as porter")
		})
	}
}

// Test that sensitive values are not included in the trace data sent to the telemetry endpoint,
// while the values of parameters that are not sensitive are.
func TestTelemetry_SensitiveValuesAreNotTraced(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// The sensitive value is used inside the bundle, so we need the traces exported from there as well
	receiver := test.StartBundleTestOTLPReceiver(tester.OTLPProtocolGRPC)

	bundleDir := filepath.Join(test.RepoRoot, "tests/integration/testdata/bundles/failing-bundle-with-sensitive-data")
	require.NoError(t, shx.Copy(filepath.Join(bundleDir, "*"), test.TestDir), "error copying the bundle into the test directory")
	test.Chdir(test.TestDir)

	// The bundle fails while running a command that has the sensitive parameter as an argument
	const sensitiveValue = "topsecret"
	_, _, err = test.RunPorter("install", "--param", "name=mybuns-author", "--param", "password="+sensitiveValue)
	require.Error(t, err, "expected the install to fail")

	// Validate that the failed install was traced, so that we know we are checking relevant trace data
	var failedSpans int
	for _, span := range receiver.RequireSpans("porter") {
		if span.GetStatus().GetCode() == tracepb.Status_STATUS_CODE_ERROR {
			failedSpans++
		}
	}
	require.NotZero(t, failedSpans, "expected the failed install to be recorded in the trace data")

	// Validate that we received trace data from the mixin that was given the sensitive value
	var mixinSpans int
	for _, span := range receiver.RequireSpans("porter") {
		if span.GetName() == "exec" {
			mixinSpans++
		}
	}
	require.NotZero(t, mixinSpans, "expected the exec mixin inside the bundle to have exported trace data")

	receiver.RequireNoSpanContains(sensitiveValue)

	// Validate that only the sensitive parameter was masked in the command recorded on the root span
	var commands []string
	for _, span := range receiver.RequireSpans("porter") {
		if span.GetName() != "porter install" {
			continue
		}
		for _, attr := range span.GetAttributes() {
			if attr.GetKey() == "command" {
				commands = append(commands, attr.GetValue().GetStringValue())
			}
		}
	}
	require.Equal(t, []string{"porter install --param name=mybuns-author --param password=*******"}, commands, "expected the root span to record the command with only the sensitive parameter masked")
}

// Test that telemetry data is being exported both from porter and the plugins
func TestTelemetry_IncludesPluginLogs(t *testing.T) {
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
	porterTrace, err := os.ReadFile(porterTraceName)
	require.NoError(t, err, "error reading porter's trace file %s", porterTraceName)
	tests.RequireOutputContains(t, string(porterTrace), `{"Key":"service.name","Value":{"Type":"STRING","Value":"porter"}}`, "no spans for porter were exported")

	// Validate we have trace data for porter
	pluginTraceName := filepath.Join(tracesDir, traces[0].Name())
	require.Contains(t, pluginTraceName, "storage.porter.mongodb", "expected the plugin trace to be for the mongodb plugin")
	pluginTrace, err := os.ReadFile(pluginTraceName)
	require.NoError(t, err, "error reading the plugin's trace file %s", pluginTraceName)
	tests.RequireOutputContains(t, string(pluginTrace), `{"Key":"service.name","Value":{"Type":"STRING","Value":"storage.porter.mongodb"}}`, "no spans for the plugins were exported")
}
