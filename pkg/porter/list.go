package porter

import (
	printer "github.com/deislabs/porter/pkg/printer"
)

// ListBundles lists bundles using the provided printer.PrintOptions
func (p *Porter) ListBundles(opts printer.PrintOptions) error {
	return p.CNAB.List(opts)
}
