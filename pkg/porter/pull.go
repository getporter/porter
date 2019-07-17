package porter

import (
	cnabtooci "github.com/deislabs/porter/pkg/cnab/cnab-to-oci"
	"github.com/pkg/errors"
)

type BundlePullOptions struct {
	Tag              string
	InsecureRegistry bool
	Force            bool
}

func (b BundlePullOptions) validateTag() error {
	_, err := cnabtooci.ParseOCIReference(b.Tag)
	if err != nil {
		return errors.Wrap(err, "invalid value for --tag, specified value should be of the form REGISTRY/bundle:tag")
	}
	return nil

}

// PullBundle looks for a given bundle tag in the bundle cache. If it is not found, it is
// pulled and stored in the cache. The path to the cached bundle is returned.
func (p *Porter) PullBundle(opts BundlePullOptions) (string, error) {
	path, ok, err := p.Cache.FindBundle(opts.Tag)
	if err != nil {
		return "", errors.Wrap(err, "unable to load bundle from cache")
	}
	// If we found the bundle, return the path to the bundle.json
	if ok && !opts.Force {
		return path, nil
	}

	b, err := p.Registry.PullBundle(opts.Tag, opts.InsecureRegistry)
	if err != nil {
		return "", errors.Wrap(err, "unable to pull remote bundle")
	}

	return p.Cache.StoreBundle(opts.Tag, b)
}
