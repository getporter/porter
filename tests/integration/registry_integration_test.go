//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	ctx := context.Background()
	c := portercontext.NewTestContext(t)
	r := cnabtooci.NewRegistry(c.Context)

	t.Run("secure registry, existing bundle", func(t *testing.T) {
		repo := cnab.MustParseOCIReference("ghcr.io/getporter/examples/porter-hello")
		ref, err := repo.WithTag("v0.2.0")
		require.NoError(t, err)

		// List Tags
		regOpts := cnabtooci.RegistryOptions{}
		tags, err := r.ListTags(ctx, repo, regOpts)
		require.NoError(t, err, "ListTags failed")
		require.Contains(t, tags, "v0.2.0", "expected a tag for the bundle version")
		require.Contains(t, tags, "3cb284ae76addb8d56b52bb7d6838351", "expected a tag for the invocation image")

		// GetBundleMetadata
		// Validates that we are passing auth when querying the registry
		meta, err := r.GetBundleMetadata(ctx, ref, regOpts)
		require.NoError(t, err, "GetBundleMetadata failed")
		require.Equal(t, "sha256:276b44be3f478b4c8d1f99c1925386d45a878a853f22436ece5589f32e9df384", meta.Digest.String(), "incorrect bundle digest")
	})

	t.Run("insecure registry, existing bundle", func(t *testing.T) {
		// Start an insecure registry with self-signed certificates
		testr, err := tester.NewTest(t)
		require.NoError(t, err, "tester.NewTest failed")
		defer testr.Close()
		reg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})

		// Copy a test bundle to the registry
		testRef := fmt.Sprintf("%s/porter-hello-nonroot:v0.2.0", reg)
		testr.RunPorter("copy", "--source=ghcr.io/getporter/examples/porter-hello:v0.2.0", "--destination", testRef, "--insecure-registry")

		// List Tags
		ref := cnab.MustParseOCIReference(testRef)
		regOpts := cnabtooci.RegistryOptions{InsecureRegistry: true}
		tags, err := r.ListTags(ctx, ref, regOpts)
		require.NoError(t, err, "ListTags failed")
		require.Contains(t, tags, "v0.2.0", "expected a tag for the bundle version")

		// GetBundleMetadata
		// Validate that call works when no auth is needed (since it's the local test registry, there is no auth)
		meta, err := r.GetBundleMetadata(ctx, ref, regOpts)
		require.NoError(t, err, "GetBundleMetadata failed")
		require.Equal(t, "sha256:276b44be3f478b4c8d1f99c1925386d45a878a853f22436ece5589f32e9df384", meta.Digest.String(), "incorrect bundle digest")
	})

	t.Run("nonexistant bundle", func(t *testing.T) {
		ref := cnab.MustParseOCIReference("ghcr.io/getporter/oops-i-dont-exist")

		// List Tags
		// Note that listing tags on a nonexistent repo will always yield an authentication error, instead of a not found, to avoid leaking information about private repositories
		regOpts := cnabtooci.RegistryOptions{}
		_, err := r.ListTags(ctx, ref, regOpts)
		require.ErrorIs(t, cnabtooci.ErrNotFound{}, err)

		// GetBundleMetadata
		// Validates that we are passing auth when querying the registry
		_, err = r.GetBundleMetadata(ctx, ref, regOpts)
		require.ErrorIs(t, cnabtooci.ErrNotFound{}, err)
	})
}
