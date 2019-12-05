package porter

import (
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
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
func (p *Porter) PullBundle(opts BundlePullOptions) (string, string, error) {
	resolver := BundleResolver{
		Cache:    p.Cache,
		Registry: p.Registry,
	}
	return resolver.Resolve(opts)
}
