package cnabprovider

import (
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
)

func (r *Runtime) LoadBundle(bundleFile string) (cnab.ExtendedBundle, error) {
	return cnab.LoadBundle(r.Context, bundleFile)
}

func (r *Runtime) ProcessBundleFromFile(bundleFile string) (cnab.ExtendedBundle, error) {
	b, err := r.LoadBundle(bundleFile)
	if err != nil {
		return cnab.ExtendedBundle{}, err
	}

	return r.ProcessBundle(b)
}

func (r *Runtime) ProcessBundle(b cnab.ExtendedBundle) (cnab.ExtendedBundle, error) {
	err := b.Validate()
	if err != nil {
		return b, fmt.Errorf("invalid bundle: %w", err)
	}

	return b, r.ProcessRequiredExtensions(b)
}
