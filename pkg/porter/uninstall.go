package porter

import (
	"fmt"
	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
)

// UninstallOptions that may be specified when uninstalling a bundle.
// Porter handles defaulting any missing values.
type UninstallOptions struct {
	sharedOptions
}

func (o *UninstallOptions) Validate(args []string) error {
	o.bundleRequired = false
	return o.sharedOptions.Validate(args)
}

// UninstallBundle accepts a set of pre-validated UninstallOptions and uses
// them to uninstall a bundle.
func (p *Porter) UninstallBundle(opts UninstallOptions) error {
	p.applyDefaultOptions(&opts.sharedOptions)

	fmt.Fprintf(p.Out, "uninstalling %s...\n", opts.Name)
	return p.Uninstall(opts.ToDuffleArgs())
}

// ToDuffleArgs converts this instance of user-provided options
// to duffle arguments.
func (o *UninstallOptions) ToDuffleArgs() cnabprovider.UninstallArguments {
	return cnabprovider.UninstallArguments{
		Claim:                 o.Name,
		BundleIdentifier:      o.File,
		BundleIsFile:          true,
		Insecure:              o.Insecure,
		Params:                o.combineParameters(),
		CredentialIdentifiers: o.CredentialIdentifiers,
	}
}
