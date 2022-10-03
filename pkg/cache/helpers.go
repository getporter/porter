package cache

import (
	"get.porter.sh/porter/pkg/cnab"
)

var _ BundleCache = &TestCache{}

// MockCache helps you test error scenarios, you don't need it for unit testing positive scenarios.
type TestCache struct {
	cache           BundleCache
	FindBundleMock  func(ref cnab.OCIReference) (CachedBundle, bool, error)
	StoreBundleMock func(bundleReference cnab.BundleReference) (CachedBundle, error)
}

func NewTestCache(cache BundleCache) *TestCache {
	return &TestCache{
		cache: cache,
	}
}

func (c *TestCache) FindBundle(ref cnab.OCIReference) (CachedBundle, bool, error) {
	if c.FindBundleMock != nil {
		return c.FindBundleMock(ref)
	}
	return c.cache.FindBundle(ref)
}

func (c *TestCache) StoreBundle(bundleRef cnab.BundleReference) (CachedBundle, error) {
	if c.StoreBundleMock != nil {
		return c.StoreBundleMock(bundleRef)
	}
	return c.cache.StoreBundle(bundleRef)
}

func (c *TestCache) GetCacheDir() (string, error) {
	return c.cache.GetCacheDir()
}
