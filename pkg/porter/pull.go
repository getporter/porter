package porter

import (
	"context"
	"fmt"

	"github.com/docker/cnab-to-oci/remotes"
	"github.com/docker/distribution/reference"
)

type BundlePullOptions struct {
	Tag              string
	InsecureRegistry bool
}

func (p *Porter) PullBundle(opts BundlePullOptions) error {

	ref, err := reference.ParseNormalizedNamed(opts.Tag)
	if err != nil {
		return err
	}

	insecureRegistries := []string{}
	if opts.InsecureRegistry {
		reg := reference.Domain(ref)
		fmt.Printf("Registry is: %s", reg)
		insecureRegistries = append(insecureRegistries, reg)
	}

	b, err := remotes.Pull(context.Background(), ref, createResolver(insecureRegistries).Resolver)
	if err != nil {
		return err
	}

	return p.writeBundle(*b)
}
