//go:build integration

package integration

import (
	"fmt"
	"path"
	"testing"

	"get.porter.sh/porter/pkg/yaml"
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

	// Try to publish with autobuild disabled, it should fail
	_, _, err = test.RunPorter("publish", "--autobuild-disabled")
	require.ErrorContains(t, err, "Skipping autobuild because --autobuild-disabled was specified")

	// Try again with autobuild disabled via a config setting instead of a flag
	// This is a regression test for https://github.com/getporter/porter/issues/2735
	test.EditYaml(path.Join(test.PorterHomeDir, "config.yaml"), func(yq *yaml.Editor) error {
		return yq.SetValue("autobuild-disabled", "true")
	})
	_, output, err := test.RunPorter("publish")
	fmt.Println(output)
	require.ErrorContains(t, err, "Skipping autobuild because --autobuild-disabled was specified")

	// Build with version override
	test.RequirePorter("build", "--version=0.0.0")

	// Start up an insecure registry with self-signed TLS certificates
	reg := test.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})
	defer reg.Close()

	// Confirm that publish picks up the version override
	// Use an insecure registry to validate that we can publish to one
	_, output = test.RequirePorter("publish", "--registry", reg.String(), "--insecure-registry")
	tests.RequireOutputContains(t, output, fmt.Sprintf("Bundle %s/porter-hello:v0.0.0 pushed successfully", reg))
}
