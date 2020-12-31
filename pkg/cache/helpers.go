package cache

import (
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
)

var _ BundleCache = &TestCache{}

// MockCache helps you test error scenarios, you don't need it for unit testing positive scenarios.
type TestCache struct {
	cache           BundleCache
	FindBundleMock  func(string) (CachedBundle, bool, error)
	StoreBundleMock func(string, bundle.Bundle, *relocation.ImageRelocationMap) (CachedBundle, error)
}

func NewTestCache(cache BundleCache) *TestCache {
	return &TestCache{
		cache: cache,
	}
}

func (c *TestCache) FindBundle(tag string) (CachedBundle, bool, error) {
	if c.FindBundleMock != nil {
		return c.FindBundleMock(tag)
	}
	return c.cache.FindBundle(tag)
}

func (c *TestCache) StoreBundle(tag string, bun bundle.Bundle, reloMap *relocation.ImageRelocationMap) (CachedBundle, error) {
	if c.StoreBundleMock != nil {
		return c.StoreBundleMock(tag, bun, reloMap)
	}
	return c.cache.StoreBundle(tag, bun, reloMap)
}

func (c *TestCache) GetCacheDir() string {
	return c.cache.GetCacheDir()
}
