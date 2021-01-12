// +build integration

package tests

import (
	"testing"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestPull_ContentDigestMissing(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	opts := porter.BundlePullOptions{}
	opts.Reference = "getporterci/mysql:no-content-digest"

	cachedBun, err := p.PullBundle(opts)
	require.Contains(t, err.Error(),
		"unable to verify that the pulled image getporterci/mysql-installer:no-content-digest is the invocation image referenced by the bundle because the bundle does not specify a content digest. This could allow for the invocation image to be replaced or tampered with")
	require.Equal(t, cache.CachedBundle{}, cachedBun)
}
