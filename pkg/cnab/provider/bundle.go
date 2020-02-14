package cnabprovider

import (
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/pkg/errors"
)

func (d *Runtime) LoadBundle(bundleFile string) (*bundle.Bundle, error) {
	l := loader.New()

	// Issue #439: Errors that come back from the loader can be
	// pretty opaque.
	bun, err := l.Load(bundleFile)
	if err != nil {
		return bun, errors.Wrapf(err, "cannot load bundle")
	}
	return bun, nil
}
