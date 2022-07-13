package porter

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/manifest"
)

// applyDefaultOptions applies more advanced defaults to the options
// based on values that beyond just what was supplied by the user
// such as information in the manifest itself.
func (p *Porter) applyDefaultOptions(ctx context.Context, opts *installationOptions) error {
	if opts.Name != "" {
		return nil
	}

	if opts.File != "" {
		m, err := manifest.LoadManifestFrom(ctx, p.Config, opts.File)
		if err != nil {
			return err
		}

		opts.Name = m.Name
		return nil
	}

	if opts.CNABFile != "" {
		bun, err := cnab.LoadBundle(p.Context, opts.CNABFile)
		if err != nil {
			return err
		}

		opts.Name = bun.Name
		return nil
	}

	return nil
}
