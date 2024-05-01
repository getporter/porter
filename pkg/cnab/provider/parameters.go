package cnabprovider

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
)

// loadParameters resolves the prepared parameter set associated with the Run, and
// updates Run.Parameters with the resolved values.
func (r *Runtime) loadParameters(ctx context.Context, b cnab.ExtendedBundle, run *storage.Run) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	resolvedParameters, err := r.parameters.ResolveAll(ctx, run.Parameters)
	if err != nil {
		return err
	}

	// Apply the resolved values back onto the run, these won't be persisted but are used in-memory
	for i, param := range run.Parameters.Parameters {
		run.Parameters.Parameters[i].ResolvedValue = resolvedParameters[param.Name]
	}

	if err = run.Parameters.ValidateBundle(b.Parameters, run.Action); err != nil {
		return span.Error(err)
	}

	if err = run.SetParametersDigest(); err != nil {
		// Just warn since the digest isn't critical for running the bundle
		// If it's not set properly, we will recalculate as needed
		span.Warnf("WARNING: unable to set the run's parameters digest: %w", err)
	}

	return nil
}
