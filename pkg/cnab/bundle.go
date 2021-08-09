package cnab

import (
	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
)

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

// ParseBundleReference into its components: repository, tag or digest
func ParseBundleReference(bundleReference string) (repo string, tag string, digest string, err error) {
	ref, err := reference.ParseNormalizedNamed(bundleReference)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "invalid bundle reference: %s", bundleReference)
	}

	repo = reference.FamiliarName(ref)

	switch v := ref.(type) {
	case reference.Tagged:
		tag = v.Tag()
	case reference.Digested:
		digest = v.Digest().String()
	}
	return repo, tag, digest, nil
}
