package cnabprovider

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
)

func (r *Runtime) loadCredentials(ctx context.Context, b cnab.ExtendedBundle, run *storage.Run) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	resolvedCredentials, err := r.credentials.ResolveAll(ctx, run.Credentials, run.Credentials.Keys())
	if err != nil {
		return span.Error(err)
	}

	for i, cred := range run.Credentials.Credentials {
		if resolvedValue, ok := resolvedCredentials[cred.Name]; ok {
			run.Credentials.Credentials[i].ResolvedValue = resolvedValue
		}
	}

	err = run.Credentials.ValidateBundle(b.Credentials, run.Action)
	if err != nil {
		return span.Error(err)
	}

	err = run.SetCredentialsDigest()
	if err != nil {
		// Just warn since the digest isn't critical for running the bundle
		// If it's not set properly, we will recalculate as needed
		span.Warnf("WARNING: unable to set the run's credentials digest: %w", err)
	}

	return nil
}
