package cnab

import (
	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

type BundleReference struct {
	Reference     OCIReference
	Digest        digest.Digest
	Definition    bundle.Bundle
	RelocationMap relocation.ImageRelocationMap
}

// LoadBundle from the specified filepath.
func LoadBundle(c *context.Context, bundleFile string) (bundle.Bundle, error) {
	bunD, err := c.FileSystem.ReadFile(bundleFile)
	if err != nil {
		return bundle.Bundle{}, errors.Wrapf(err, "cannot read bundle at %s", bundleFile)
	}

	bun, err := bundle.Unmarshal(bunD)
	if err != nil {
		return bundle.Bundle{}, errors.Wrapf(err, "cannot load bundle from\n%s at %s", string(bunD), bundleFile)
	}

	return *bun, nil
}
