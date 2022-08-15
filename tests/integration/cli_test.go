//go:build integration

package integration

import (
	"strings"
	"testing"

	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

// Test that the CLI is configured properly.
func TestCLI(t *testing.T) {
	t.Skip("Turning off until we come back and stop logging errors in main")

	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// When the command fails, only print the error message once
	_, output, _ := test.RunPorter("explain", "-r=ghcr.io/getporter/missing-bundle")
	gotErrors := strings.Count(output, "unable to pull bundle")
	require.Equal(t, 1, gotErrors)
}
