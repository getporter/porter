//go:build integration

package cnabtooci

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/require"
)

func TestRegistry_ListTags(t *testing.T) {
	ctx := context.Background()
	c := portercontext.NewTestContext(t)
	r := NewRegistry(c.Context)

	t.Run("existing bundle", func(t *testing.T) {
		tags, err := r.ListTags(ctx, "docker.io/carolynvs/porter-hello-nonroot")
		require.NoError(t, err, "ListTags failed")

		require.Contains(t, tags, "v0.1.0", "expected a tag for the bundle version")
		require.Contains(t, tags, "0540db3f2c70103816cc91e9c4207447", "expected a tag for the invocation image")
	})

	t.Run("nonexistant bundle", func(t *testing.T) {
		_, err := r.ListTags(ctx, "docker.io/carolynvs/oops-i-dont-exist")
		tests.RequireErrorContains(t, err, "error listing tags for docker.io/carolynvs/oops-i-dont-exist")
		// Note that listing tags on a nonexistent repo will always yield an authentication error, instead of a not found, to avoid leaking information about private repositories
	})
}
