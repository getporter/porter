package v2

import (
	"context"
)

var _ DependencyResolver = DefaultBundleResolver{}

// DefaultBundleResolver resolves the default bundle defined on the dependency.
type DefaultBundleResolver struct {
	puller BundlePuller
}

func (d DefaultBundleResolver) ResolveDependency(ctx context.Context, dep Dependency) (Node, bool, error) {
	if dep.DefaultBundle == nil {
		return nil, false, nil
	}

	cb, err := d.puller.GetBundle(ctx, dep.DefaultBundle.Reference)
	if err != nil {
		// wrap not found error and indicate that we could resolve anything
		return nil, false, err
	}

	return BundleNode{
		Key:       dep.Key,
		ParentKey: dep.ParentKey,
		Reference: cb.BundleReference,
		// TODO(PEP003): Do we have to duplicate this? Can't we get it from the bundle def when we need it?
		Parameters:  dep.Parameters,
		Credentials: dep.Credentials,
	}, true, nil
}
