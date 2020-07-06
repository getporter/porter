package cnabprovider

import (
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/pkg/errors"
)

func (r *Runtime) LoadBundle(bundleFile string) (bundle.Bundle, error) {
	l := loader.New()

	bunD, err := r.FileSystem.ReadFile(bundleFile)
	if err != nil {
		return bundle.Bundle{}, errors.Wrapf(err, "cannot read bundle at %s", bundleFile)
	}

	// Issue #439: Errors that come back from the loader can be
	// pretty opaque.
	bun, err := l.LoadData(bunD)
	if err != nil {
		return bundle.Bundle{}, errors.Wrapf(err, "cannot load bundle")
	}

	return *bun, nil
}

func (r *Runtime) ProcessBundle(bundleFile string) (bundle.Bundle, error) {
	b, err := r.LoadBundle(bundleFile)
	if err != nil {
		return bundle.Bundle{}, err
	}

	err = b.Validate()
	if err != nil {
		return b, errors.Wrap(err, "invalid bundle")
	}

	return b, r.ProcessRequiredExtensions(b)
}
