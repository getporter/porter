//go:build integration

package integration

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
	"github.com/uwu-tools/magex/shx"
)

func TestPluginDebugLogsVerbosityArgument(t *testing.T) {
	testcases := []struct {
		name              string
		verbosity         string
		shouldContain     bool
		expectedPluginLog string
	}{
		{"plugin debug logs", "debug", true, "plugin started"},
		{"plugin info logs doesn't contain plugin process exited", "info", false, "plugin process exited"},
		{"plugin debug logs contains plugin process exited", "debug", true, "plugin process exited"},
		{"plugin info logs", "info", false, "plugin started"},
	}
	for _, tc := range testcases {
		test, err := tester.NewTest(t)
		require.NoError(t, err, "test setup failed")
		defer test.Close()

		output, _ := test.RequirePorter("list", "--verbosity", tc.verbosity)
		if tc.shouldContain {
			require.Contains(t, output, tc.expectedPluginLog)
		} else {
			require.NotContains(t, output, tc.expectedPluginLog)
		}
	}
}

func TestPluginDebugLogsVerbosityEnvironmentVariable(t *testing.T) {
	testcases := []struct {
		name              string
		verbosity         string
		shouldContain     bool
		expectedPluginLog string
	}{
		{"plugin debug logs", "debug", true, "plugin started"},
		{"plugin info logs doesn't contain plugin process exited", "info", false, "plugin process exited"},
		{"plugin debug logs contains plugin process exited", "debug", true, "plugin process exited"},
		{"plugin info logs", "info", false, "plugin started"},
	}
	for _, tc := range testcases {
		test, err := tester.NewTest(t)
		require.NoError(t, err, "test setup failed")
		defer test.Close()

		output, _, err := test.RunPorterWith(func(cmd *shx.PreparedCommand) {
			cmd.Args("list")
			cmd.Env(fmt.Sprintf("PORTER_VERBOSITY=%s", tc.verbosity))
		})
		if tc.shouldContain {
			require.Contains(t, output, tc.expectedPluginLog)
		} else {
			require.NotContains(t, output, tc.expectedPluginLog)
		}
	}
}
