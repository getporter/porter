package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/cache"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
)

func TestBundleResolver_Resolve_ForcePull(t *testing.T) {
	tc := config.NewTestConfig(t)
	testReg := cnabtooci.NewTestRegistry()
	testCache := cache.NewTestCache(cache.New(tc.Config))
	resolver := BundleResolver{
		Cache:    testCache,
		Registry: testReg,
	}

	cacheSearched := false
	testCache.FindBundleMock = func(tag string) (cache.CachedBundle, bool, error) {
		cacheSearched = true
		return cache.CachedBundle{}, true, nil
	}

	pulled := false
	testReg.MockPullBundle = func(tag string, insecureRegistry bool) (bundle.Bundle, *relocation.ImageRelocationMap, error) {
		pulled = true
		return bundle.Bundle{}, nil, nil
	}

	opts := BundlePullOptions{
		Force: true,
	}
	resolver.Resolve(opts)

	assert.False(t, cacheSearched, "Force should have skipped the cache")
	assert.True(t, pulled, "The bundle should have been re-pulled")
}

func TestBundleResolver_Resolve_CacheHit(t *testing.T) {
	tc := config.NewTestConfig(t)
	testReg := cnabtooci.NewTestRegistry()
	testCache := cache.NewTestCache(cache.New(tc.Config))
	resolver := BundleResolver{
		Cache:    testCache,
		Registry: testReg,
	}

	cacheSearched := false
	testCache.FindBundleMock = func(tag string) (cache.CachedBundle, bool, error) {
		cacheSearched = true
		return cache.CachedBundle{}, true, nil
	}

	pulled := false
	testReg.MockPullBundle = func(tag string, insecureRegistry bool) (bundle.Bundle, *relocation.ImageRelocationMap, error) {
		pulled = true
		return bundle.Bundle{}, nil, nil
	}

	opts := BundlePullOptions{}
	resolver.Resolve(opts)

	assert.True(t, cacheSearched, "The cache should be searched when force is not specified")
	assert.False(t, pulled, "The bundle should NOT be pulled because it was found in the cache")
}

func TestBundleResolver_Resolve_CacheMiss(t *testing.T) {
	tc := config.NewTestConfig(t)
	testReg := cnabtooci.NewTestRegistry()
	testCache := cache.NewTestCache(cache.New(tc.Config))
	resolver := BundleResolver{
		Cache:    testCache,
		Registry: testReg,
	}

	cacheSearched := false
	testCache.FindBundleMock = func(tag string) (cache.CachedBundle, bool, error) {
		cacheSearched = true
		return cache.CachedBundle{}, false, nil
	}

	pulled := false
	testReg.MockPullBundle = func(tag string, insecureRegistry bool) (bundle.Bundle, *relocation.ImageRelocationMap, error) {
		pulled = true
		return bundle.Bundle{}, nil, nil
	}

	opts := BundlePullOptions{}
	resolver.Resolve(opts)

	assert.True(t, cacheSearched, "The cache should be searched when force is not specified")
	assert.True(t, pulled, "The bundle should have been pulled because the bundle was not in the cache")
}
