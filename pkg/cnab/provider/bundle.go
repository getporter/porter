package cnabprovider

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
)

func (r *Runtime) LoadBundle(bundleFile string) (cnab.ExtendedBundle, error) {
	return cnab.LoadBundle(r.Context, bundleFile)
}

func (r *Runtime) ProcessBundleFromFile(ctx context.Context, bundleFile string) (cnab.ExtendedBundle, error) {
	b, err := r.LoadBundle(bundleFile)
	if err != nil {
		return cnab.ExtendedBundle{}, err
	}

	return r.ProcessBundle(ctx, b)
}

func (r *Runtime) ProcessBundle(ctx context.Context, b cnab.ExtendedBundle) (cnab.ExtendedBundle, error) {
	strategy := r.GetSchemaCheckStrategy(ctx)
	err := b.Validate(r.Context, strategy)
	if err != nil {
		return b, fmt.Errorf("invalid bundle: %w", err)
	}

	return b, r.ProcessRequiredExtensions(b)
}
