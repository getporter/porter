package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/cnab"
)

type BundlePullOptions struct {
	Reference        string
	_ref             *cnab.OCIReference
	InsecureRegistry bool
	Force            bool
}

func (b *BundlePullOptions) Validate() error {
	return b.validateReference()
}

func (b *BundlePullOptions) GetReference() cnab.OCIReference {
	if b._ref == nil {
		ref, err := cnab.ParseOCIReference(b.Reference)
		if err != nil {
			panic(err)
		}
		b._ref = &ref
	}
	return *b._ref
}

func (b *BundlePullOptions) validateReference() error {
	ref, err := cnab.ParseOCIReference(b.Reference)
	if err != nil {
		return fmt.Errorf("invalid value for --reference, specified value should be of the form REGISTRY/bundle:tag: %w", err)
	}
	b._ref = &ref
	return nil
}

// PullBundle looks for a given bundle tag in the bundle cache. If it is not found, it is
// pulled and stored in the cache. The path to the cached bundle is returned.
func (p *Porter) PullBundle(ctx context.Context, opts BundlePullOptions) (cache.CachedBundle, error) {
	resolver := BundleResolver{
		Cache:    p.Cache,
		Registry: p.Registry,
	}
	return resolver.Resolve(ctx, opts)
}
