package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	depsv2 "get.porter.sh/porter/pkg/cnab/dependencies/v2"

	"get.porter.sh/porter/pkg/cache"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/tracing"
)

var _ depsv2.BundlePuller = BundleResolver{}

// BundleResolver supports retrieving bundles from a registry, with cache support.
type BundleResolver struct {
	cache    cache.BundleCache
	registry cnabtooci.RegistryProvider
	regOpts  cnabtooci.RegistryOptions

	// refreshCache always pulls from the registry, and ignores cached bundles.
	refreshCache bool
}

func NewBundleResolver(cache cache.BundleCache, refreshCache bool, registry cnabtooci.RegistryProvider, regOpts cnabtooci.RegistryOptions) BundleResolver {
	return BundleResolver{
		cache:        cache,
		refreshCache: refreshCache,
		registry:     registry,
		regOpts:      regOpts,
	}
}

func (r BundleResolver) GetBundle(ctx context.Context, ref cnab.OCIReference) (cache.CachedBundle, error) {
	log := tracing.LoggerFromContext(ctx)

	if !r.refreshCache {
		cachedBundle, ok, err := r.cache.FindBundle(ref)
		if err != nil {
			return cache.CachedBundle{}, log.Error(fmt.Errorf("unable to load bundle %s from cache: %w", ref, err))
		}
		// If we found the bundle, return the path to the bundle.json
		if ok {
			return cachedBundle, nil
		}
	}

	bundleRef, err := r.registry.PullBundle(ctx, ref, r.regOpts)
	if err != nil {
		return cache.CachedBundle{}, err
	}

	cb, err := r.cache.StoreBundle(bundleRef)
	if err != nil {
		return cache.CachedBundle{}, log.Errorf("error storing the bundle %s in the Porter bundle cache: %w", bundleRef, err)
	}
	return cb, nil
}

func (r BundleResolver) ListTags(ctx context.Context, repo cnab.OCIReference) ([]string, error) {
	return r.registry.ListTags(ctx, repo, r.regOpts)
}
