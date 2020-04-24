package porter

import (
	"get.porter.sh/porter/pkg/cache"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"github.com/pkg/errors"
)

type BundleResolver struct {
	Cache    cache.BundleCache
	Registry cnabtooci.RegistryProvider
}

// Resolves a bundle from the cache, or pulls it and caches it
// Returns the location of the bundle or an error
func (r *BundleResolver) Resolve(opts BundlePullOptions) (cache.CachedBundle, error) {
	if !opts.Force {
		cachedBundle, ok, err := r.Cache.FindBundle(opts.Tag)
		if err != nil {
			return cache.CachedBundle{}, errors.Wrapf(err, "unable to load bundle %s from cache", opts.Tag)
		}
		// If we found the bundle, return the path to the bundle.json
		if ok {
			return cachedBundle, nil
		}
	}

	b, rMap, err := r.Registry.PullBundle(opts.Tag, opts.InsecureRegistry)
	if err != nil {
		return cache.CachedBundle{}, err
	}

	return r.Cache.StoreBundle(opts.Tag, b, rMap)
}
