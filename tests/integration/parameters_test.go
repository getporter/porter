//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

func TestParametersOutputContainsParameterSetName(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")
	test.TestContext.AddTestFile("testdata/parametersets/params.json", "params.json")
	test.RequirePorter("parameters", "apply", "params.json")

	output, _ := test.RequirePorter("parameters", "list")
	require.Contains(t, output, "integration-test")
}
