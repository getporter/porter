package cnabprovider

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
)

// loadCredentials resolves prepared credential set associated with the Run, and
// updates Run.Credentials with the resolved values.
func (r *Runtime) loadCredentials(ctx context.Context, b cnab.ExtendedBundle, run *storage.Run) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	resolvedCredentials, err := r.credentials.ResolveAll(ctx, run.Credentials)
	if err != nil {
		return err
	}

	// Apply the resolved values back onto the run, these won't be persisted but are used in-memory
	for i, cred := range run.Credentials.Credentials {
		run.Credentials.Credentials[i].Value = resolvedCredentials[cred.Name]
	}

	if err = run.Credentials.ValidateBundle(b.Credentials, run.Action); err != nil {
		return span.Error(err)
	}

	if err = run.SetCredentialsDigest(); err != nil {
		// Just warn since the digest isn't critical for running the bundle
		// If it's not set properly, we will recalculate as needed
		span.Warnf("WARNING: unable to set the run's credentials digest: %w", err)
	}

	return nil
}
