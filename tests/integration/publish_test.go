//go:build integration

package integration

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

func TestPublish(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Create a bundle
	test.Chdir(test.TestDir)
	test.RequirePorter("create")

	// Build with version override
	test.RequirePorter("build", "--version=0.0.0")

	// Start up an insecure registry with self-signed TLS certificates
	reg := test.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})

	// Confirm that publish picks up the version override
	// Use an insecure registry to validate that we can publish to one
	_, output := test.RequirePorter("publish", "--registry", reg.String(), "--insecure-registry")
	tests.RequireOutputContains(t, output, fmt.Sprintf("Bundle %s/porter-hello:v0.0.0 pushed successfully", reg))
}
