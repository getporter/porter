package porter

import (
	"github.com/deislabs/cnab-go/bundle"
	"github.com/pkg/errors"
)

type BundleLifecycleOpts struct {
	sharedOptions
	BundlePullOptions
}

// prepullBundleByTag handles calling the bundle pull operation and updating
// the shared options like name and bundle file path. This is used by install, upgrade
// and uninstall
func (p *Porter) prepullBundleByTag(opts *BundleLifecycleOpts) error {
	if opts.Tag == "" {
		return nil
	}

	bundlePath, err := p.PullBundle(opts.BundlePullOptions)
	if err != nil {
		return errors.Wrapf(err, "unable to pull bundle %s", opts.Tag)
	}
	opts.CNABFile = bundlePath
	rdr, err := p.Config.FileSystem.Open(bundlePath)
	if err != nil {
		return errors.Wrap(err, "unable to open bundle file")
	}
	defer rdr.Close()
	bun, err := bundle.ParseReader(rdr)
	if opts.Name == "" {
		opts.Name = bun.Name
	}
	return nil
}
