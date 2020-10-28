package cnab

import (
	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/pkg/errors"
)

func IsFileType(s *definition.Schema) bool {
	return s.Type == "string" && s.ContentEncoding == "base64"
}

func LoadBundle(c *context.Context, bundleFile string) (bundle.Bundle, error) {
	l := loader.New()

	bunD, err := c.FileSystem.ReadFile(bundleFile)
	if err != nil {
		return bundle.Bundle{}, errors.Wrapf(err, "cannot read bundle at %s", bundleFile)
	}

	bun, err := l.LoadData(bunD)
	if err != nil {
		return bundle.Bundle{}, errors.Wrapf(err, "cannot load bundle from\n%s at %s", string(bunD), bundleFile)
	}

	return *bun, nil
}
