package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleResolver_Resolve_ForcePull(t *testing.T) {
	ctx := context.Background()
	tc := config.NewTestConfig(t)
	testReg := cnabtooci.NewTestRegistry()
	testCache := cache.NewTestCache(cache.New(tc.Config))
	resolver := BundleResolver{
		Cache:    testCache,
		Registry: testReg,
	}

	cacheSearched := false
	testCache.FindBundleMock = func(ref cnab.OCIReference) (cache.CachedBundle, bool, error) {
		cacheSearched = true
		return cache.CachedBundle{}, true, nil
	}

	pulled := false
	testReg.MockPullBundle = func(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
		pulled = true
		return cnab.BundleReference{Reference: ref}, nil
	}

	opts := BundlePullOptions{
		Reference: kahnlatest.String(),
		Force:     true,
	}
	require.NoError(t, opts.Validate())
	resolver.Resolve(ctx, opts)

	assert.False(t, cacheSearched, "Force should have skipped the cache")
	assert.True(t, pulled, "The bundle should have been re-pulled")
}

func TestBundleResolver_Resolve_CacheHit(t *testing.T) {
	ctx := context.Background()
	tc := config.NewTestConfig(t)
	testReg := cnabtooci.NewTestRegistry()
	testCache := cache.NewTestCache(cache.New(tc.Config))
	resolver := BundleResolver{
		Cache:    testCache,
		Registry: testReg,
	}

	cacheSearched := false
	testCache.FindBundleMock = func(ref cnab.OCIReference) (cache.CachedBundle, bool, error) {
		cacheSearched = true
		return cache.CachedBundle{BundleReference: cnab.BundleReference{Reference: ref}}, true, nil
	}

	pulled := false
	testReg.MockPullBundle = func(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
		pulled = true
		return cnab.BundleReference{Reference: ref}, nil
	}

	opts := BundlePullOptions{Reference: "ghcr.io/getporter/examples/porter-hello:v0.2.0"}
	resolver.Resolve(ctx, opts)

	assert.True(t, cacheSearched, "The cache should be searched when force is not specified")
	assert.False(t, pulled, "The bundle should NOT be pulled because it was found in the cache")
}

func TestBundleResolver_Resolve_CacheMiss(t *testing.T) {
	ctx := context.Background()
	tc := config.NewTestConfig(t)
	testReg := cnabtooci.NewTestRegistry()
	testCache := cache.NewTestCache(cache.New(tc.Config))
	resolver := BundleResolver{
		Cache:    testCache,
		Registry: testReg,
	}

	cacheSearched := false
	testCache.FindBundleMock = func(ref cnab.OCIReference) (cache.CachedBundle, bool, error) {
		cacheSearched = true
		return cache.CachedBundle{}, false, nil
	}

	pulled := false
	testReg.MockPullBundle = func(ctx context.Context, ref cnab.OCIReference, options cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
		pulled = true
		return cnab.BundleReference{Reference: ref}, nil
	}

	opts := BundlePullOptions{Reference: "ghcr.io/getporter/examples/porter-hello:v0.2.0"}
	resolver.Resolve(ctx, opts)

	assert.True(t, cacheSearched, "The cache should be searched when force is not specified")
	assert.True(t, pulled, "The bundle should have been pulled because the bundle was not in the cache")
}
