package porter

import (
	"get.porter.sh/porter/pkg/manifest"
)

// applyDefaultOptions applies more advanced defaults to the options
// based on values that beyond just what was supplied by the user
// such as information in the manifest itself.
func (p *Porter) applyDefaultOptions(opts *sharedOptions) error {
	if opts.File != "" {
		err := p.LoadManifestFrom(opts.File)
		if err != nil {
			return err
		}
	}

	// Ensure that we have a manifest initialized, even if it's just an empty one
	// This happens for non-porter bundles using --cnab-file or --tag
	if p.Manifest == nil {
		p.Manifest = &manifest.Manifest{}
	}

	//
	// Default the installation name to the bundle name
	//
	if opts.Name == "" {
		opts.Name = p.Manifest.Name
	}

	return nil
}
