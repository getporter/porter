//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"path"
	"testing"

	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/google/go-containerregistry/pkg/crane"
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
	_, output, _ := test.RunPorter("publish", "--autobuild-disabled")
	tests.RequireOutputContains(t, output, "Skipping autobuild because --autobuild-disabled was specified")

	// Try again with autobuild disabled via a config setting instead of a flag
	// This is a regression test for https://github.com/getporter/porter/issues/2735
	test.EditYaml(path.Join(test.PorterHomeDir, "config.yaml"), func(yq *yaml.Editor) error {
		return yq.SetValue("autobuild-disabled", "true")
	})
	_, output, _ = test.RunPorter("publish")
	fmt.Println(output)
	tests.RequireOutputContains(t, output, "Skipping autobuild because --autobuild-disabled was specified")

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

func TestPublish_PreserveTags(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Start a temporary registry, that uses plain http (no TLS)
	reg := test.StartTestRegistry(tester.TestRegistryOptions{UseTLS: false})
	defer reg.Close()

	ref := fmt.Sprintf("%s/embeddedimg:v0.1.1", reg)
	test.MakeTestBundle(testdata.EmbeddedImg, ref, tester.PreserveTags)

	taggedDigest, err := crane.Digest(fmt.Sprintf("%s/alpine:3.20.3", reg), crane.Insecure)
	require.NoError(t, err)

	// Confirm that the digest is the same
	output, _ := test.RequirePorter("inspect", ref, "-o", "json", "--verbosity", "info")
	var images struct {
		Images []struct {
			Digest string `json:"contentDigest"`
		} `json:"images"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &images))
	require.Equal(t, 1, len(images.Images))
	require.Equal(t, taggedDigest, images.Images[0].Digest)
}

func TestPublish_PreserveTagsChanged(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Start a temporary registry, that uses plain http (no TLS)
	reg := test.StartTestRegistry(tester.TestRegistryOptions{UseTLS: false})
	defer reg.Close()

	ref := fmt.Sprintf("%s/embeddedimg:v0.1.1", reg)
	test.MakeTestBundle(testdata.EmbeddedImg, ref)
	_, err = crane.Digest(fmt.Sprintf("%s/alpine:3.20.3", reg), crane.Insecure)
	require.Error(t, err)

	ref = fmt.Sprintf("%s/embeddedimg:v0.1.2", reg)
	test.MakeTestBundle(testdata.EmbeddedImg, ref, tester.PreserveTags)
	_, err = crane.Digest(fmt.Sprintf("%s/alpine:3.20.3", reg), crane.Insecure)
	require.NoError(t, err)
}
