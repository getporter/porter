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
)

func TestSyft(t *testing.T) {
	test, err := tester.NewTestWithConfig(t, "tests/integration/testdata/porter-sbom-config.yaml")
	require.NoError(t, err, "tester.NewTest failed")
	defer test.Close()

	// Create a bundle
	test.Chdir(test.TestDir)
	test.RequirePorter("create")

	// Build with version override
	test.RequirePorter("build", "--version=0.0.0")

	// Start up an insecure registry with self-signed TLS certificates
	reg := test.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})
	defer reg.Close()

	// Confirm that publish picks up the version override
	// Use an insecure registry to validate that we can publish to one
	sbomFilePath := filepath.Join(test.TestDir, "sbom.json")
	_, output := test.RequirePorter("publish", "--registry", reg.String(), "--insecure-registry", "--sbom-file", sbomFilePath)
	tests.RequireOutputContains(t, output, fmt.Sprintf("Bundle %s/porter-hello:v0.0.0 pushed successfully", reg))

	sbomContent, err := os.ReadFile(sbomFilePath)
	require.NoError(t, err, "error reading the sbom file %s", sbomFilePath)
	tests.RequireOutputContains(t, string(sbomContent), `"spdxVersion":"SPDX-`, "no SPDX version found in sbom content")
	tests.RequireOutputContains(t, string(sbomContent), `/porter-hello"`, "no image tag found in sbom content")
}
