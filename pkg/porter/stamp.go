package porter

import (
	"fmt"

	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"github.com/deislabs/cnab-go/bundle"
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
		return p.Build(BuildOptions{})
	}
	return nil
}

// IsBundleUpToDate checks the hash of the manifest against the hash in cnab/bundle.json.
func (p *Porter) IsBundleUpToDate(opts bundleFileOptions) (bool, error) {
	if exists, _ := p.FileSystem.Exists(opts.CNABFile); exists {
		bunData, err := p.FileSystem.ReadFile(opts.CNABFile)
		if err != nil {
			return false, errors.Wrapf(err, "could not read data from %s", opts.CNABFile)
		}

		bun, err := bundle.Unmarshal(bunData)
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

		converter := configadapter.NewManifestConverter(p.Context, p.Manifest, nil, mixins)
		newStamp := converter.GenerateStamp()
		return oldStamp.ManifestDigest == newStamp.ManifestDigest, nil
	}

	return false, nil
}
