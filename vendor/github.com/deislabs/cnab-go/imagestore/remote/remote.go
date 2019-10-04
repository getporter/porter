package remote

import (
	"fmt"

	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
	"github.com/pivotal/image-relocation/pkg/registry/ggcr"

	"github.com/deislabs/cnab-go/imagestore"
)

// remote is an image store which does not actually store images. It is used to represent thin bundles.
type remote struct {
	registryClient registry.Client
}

func Create(...imagestore.Option) (imagestore.Store, error) {
	return &remote{
		registryClient: ggcr.NewRegistryClient(),
	}, nil
}

func (r *remote) Add(im string) (string, error) {
	return "", nil
}

func (r *remote) Push(d image.Digest, src image.Name, dst image.Name) error {
	dig, _, err := r.registryClient.Copy(src, dst)
	if err != nil {
		return err
	}

	if d != image.EmptyDigest && dig != d {
		return fmt.Errorf("digest of image %s not preserved: old digest %s; new digest %s", src, d.String(), dig.String())
	}
	return nil
}
