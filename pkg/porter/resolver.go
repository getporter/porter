package porter

import (
	"github.com/deislabs/porter/pkg/cache"
	"github.com/pkg/errors"
)

type BundleResolver struct {
	Cache    cache.BundleCache
	Registry Registry
}

// Resolves a bundle from the cache, or pulls it and caches it
// Returns the location of the bundle or an error
func (r *BundleResolver) Resolve(opts BundlePullOptions) (string, error) {
	path, ok, err := r.Cache.FindBundle(opts.Tag)
	if err != nil {
		return "", errors.Wrapf(err, "unable to load bundle %s from cache", opts.Tag)
	}
	// If we found the bundle, return the path to the bundle.json
	if ok && !opts.Force {
		return path, nil
	}

	b, _, err := r.Registry.PullBundle(opts.Tag, opts.InsecureRegistry)
	if err != nil {
		return "", err
	}

	return r.Cache.StoreBundle(opts.Tag, b)
}
