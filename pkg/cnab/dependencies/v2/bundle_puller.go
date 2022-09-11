package v2

import (
	"context"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/cnab"
)

// BundlePuller can query and pull bundles.
type BundlePuller interface {
	// GetBundle retrieves a bundle definition.
	GetBundle(ctx context.Context, ref cnab.OCIReference) (cache.CachedBundle, error)

	// ListTags retrieves all tags defined for a bundle.
	ListTags(ctx context.Context, ref cnab.OCIReference) ([]string, error)
}
