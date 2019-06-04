package porter

import (
	"context"

	"github.com/docker/cnab-to-oci/remotes"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
)

type BundlePullOptions struct {
	Tag              string
	InsecureRegistry bool
	Force            bool
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
	ref, err := reference.ParseNormalizedNamed(opts.Tag)
	if err != nil {
		return "", errors.Wrap(err, "invalid bundle tag format, expected REGISTRY/name:tag")
	}

	insecureRegistries := []string{}
	if opts.InsecureRegistry {
		reg := reference.Domain(ref)
		insecureRegistries = append(insecureRegistries, reg)
	}

	b, err := remotes.Pull(context.Background(), ref, p.createResolver(insecureRegistries).Resolver)
	if err != nil {
		return "", errors.Wrap(err, "unable to pull remote bundle")
	}

	return p.Cache.StoreBundle(opts.Tag, b)
}
