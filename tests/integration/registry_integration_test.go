//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

func TestRegistry_ListTags(t *testing.T) {
	ctx := context.Background()
	c := portercontext.NewTestContext(t)
	r := cnabtooci.NewRegistry(c.Context)

	t.Run("secure registry, existing bundle", func(t *testing.T) {
		ref := cnab.MustParseOCIReference("docker.io/carolynvs/porter-hello-nonroot")
		tags, err := r.ListTags(ctx, ref, cnabtooci.RegistryOptions{})
		require.NoError(t, err, "ListTags failed")

		require.Contains(t, tags, "v0.1.0", "expected a tag for the bundle version")
		require.Contains(t, tags, "0540db3f2c70103816cc91e9c4207447", "expected a tag for the invocation image")
	})

	t.Run("insecure registry, existing bundle", func(t *testing.T) {
		// Start an insecure registry with self-signed certificates
		testr, err := tester.NewTest(t)
		require.NoError(t, err, "tester.NewTest failed")
		defer testr.Close()
		reg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})

		// Copy a test bundle to the registry
		testRef := fmt.Sprintf("%s/porter-hello-nonroot:v0.1.0", reg)
		testr.RunPorter("copy", "--source=docker.io/carolynvs/porter-hello-nonroot:v0.1.0", "--destination", testRef, "--insecure-registry")

		ref := cnab.MustParseOCIReference(testRef)
		tags, err := r.ListTags(ctx, ref, cnabtooci.RegistryOptions{InsecureRegistry: true})
		require.NoError(t, err, "ListTags failed")

		require.Contains(t, tags, "v0.1.0", "expected a tag for the bundle version")
	})

	t.Run("nonexistant bundle", func(t *testing.T) {
		ref := cnab.MustParseOCIReference("docker.io/carolynvs/oops-i-dont-exist")
		_, err := r.ListTags(ctx, ref, cnabtooci.RegistryOptions{})
		tests.RequireErrorContains(t, err, "error listing tags for docker.io/carolynvs/oops-i-dont-exist")
		// Note that listing tags on a nonexistent repo will always yield an authentication error, instead of a not found, to avoid leaking information about private repositories
	})
}
