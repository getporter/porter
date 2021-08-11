package cnabprovider

import (
	"get.porter.sh/porter/pkg/cnab"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

func (r *Runtime) LoadBundle(bundleFile string) (bundle.Bundle, error) {
	return cnab.LoadBundle(r.Context, bundleFile)
}

func (r *Runtime) ProcessBundleFromFile(bundleFile string) (bundle.Bundle, error) {
	b, err := r.LoadBundle(bundleFile)
	if err != nil {
		return bundle.Bundle{}, err
	}

	return r.ProcessBundle(b)
}

func (r *Runtime) ProcessBundle(b bundle.Bundle) (bundle.Bundle, error) {
	err := b.Validate()
	if err != nil {
		return b, errors.Wrap(err, "invalid bundle")
	}

	return b, r.ProcessRequiredExtensions(b)
}
