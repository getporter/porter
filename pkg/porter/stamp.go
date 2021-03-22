package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"github.com/pkg/errors"
)

// ensureLocalBundleIsUpToDate ensures that the bundle is up to date with the porter manifest,
// if it is out-of-date, performs a build of the bundle.
func (p *Porter) ensureLocalBundleIsUpToDate(opts bundleFileOptions) error {
	if opts.File == "" {
		return nil
	}

	upToDate, err := p.IsBundleUpToDate(opts)
	if err != nil {
		fmt.Fprintln(p.Err, "warning", err)
	}

	if !upToDate {
		fmt.Fprintln(p.Out, "Building bundle ===>")
		// opts.File is non-empty, which overrides opts.CNABFile if set
		// (which may be if a cached bundle is fetched e.g. when running an action)
		opts.CNABFile = ""
		return p.Build(BuildOptions{bundleFileOptions: opts})
	}
	return nil
}

// IsBundleUpToDate checks the hash of the manifest against the hash in cnab/bundle.json.
func (p *Porter) IsBundleUpToDate(opts bundleFileOptions) (bool, error) {
	if exists, _ := p.FileSystem.Exists(opts.CNABFile); exists {
		bun, err := cnab.LoadBundle(p.Context, opts.CNABFile)
		if err != nil {
			return false, errors.Wrapf(err, "could not marshal data from %s", opts.CNABFile)
		}

		oldStamp, err := configadapter.LoadStamp(bun)
		if err != nil {
			return false, errors.Wrapf(err, "could not load stamp from %s", opts.CNABFile)
		}

		mixins, err := p.getUsedMixins()
		if err != nil {
			return false, errors.Wrapf(err, "error while listing used mixins")
		}

		converter := configadapter.NewManifestConverter(p.Context, p.Manifest, opts.File, nil, mixins)
		newDigest, err := converter.DigestManifest()
		if err != nil {
			if p.Debug {
				fmt.Fprintln(p.Err, errors.Wrap(err, "could not determine if the bundle is up-to-date so will rebuild just in case"))
			}
			return false, nil
		}
		return oldStamp.ManifestDigest == newDigest, nil
	}

	return false, nil
}
