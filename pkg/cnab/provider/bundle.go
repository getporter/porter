package cnabprovider

import (
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/pkg/errors"
)

func (d *Runtime) LoadBundle(bundleFile string) (*bundle.Bundle, error) {
	l := loader.New()

	bunD, err := d.FileSystem.ReadFile(bundleFile)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read bundle at %s", bundleFile)
	}

	// Issue #439: Errors that come back from the loader can be
	// pretty opaque.
	bun, err := l.LoadData(bunD)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load bundle")
	}

	return bun, nil
}
