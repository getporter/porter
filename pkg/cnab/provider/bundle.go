package cnabprovider

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/tracing"
)

func (r *Runtime) LoadBundle(bundleFile string) (cnab.ExtendedBundle, error) {
	return cnab.LoadBundle(r.Context, bundleFile)
}

func (r *Runtime) ProcessBundleFromFile(ctx context.Context, bundleFile string) (cnab.ExtendedBundle, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	b, err := r.LoadBundle(bundleFile)
	if err != nil {
		return cnab.ExtendedBundle{}, span.Error(err)
	}

	return r.ProcessBundle(ctx, b)
}

func (r *Runtime) ProcessBundle(ctx context.Context, b cnab.ExtendedBundle) (cnab.ExtendedBundle, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	strategy := r.GetSchemaCheckStrategy(ctx)
	err := b.Validate(r.Context, strategy)
	if err != nil {
		return b, span.Errorf("invalid bundle: %w", err)
	}

	return b, span.Error(r.ProcessRequiredExtensions(b))
}
