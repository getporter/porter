package porter

import (
	"fmt"

	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/context"
)

// UninstallOptions that may be specified when uninstalling a bundle.
// Porter handles defaulting any missing values.
type UninstallOptions struct {
	sharedOptions
}

func (o *UninstallOptions) Validate(args []string, cxt *context.Context) error {
	return o.sharedOptions.Validate(args, cxt)
}

// UninstallBundle accepts a set of pre-validated UninstallOptions and uses
// them to uninstall a bundle.
func (p *Porter) UninstallBundle(opts UninstallOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	err = p.EnsureBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "uninstalling %s...\n", opts.Name)
	return p.CNAB.Uninstall(opts.ToDuffleArgs())
}

// ToDuffleArgs converts this instance of user-provided options
// to duffle arguments.
func (o *UninstallOptions) ToDuffleArgs() cnabprovider.UninstallArguments {
	return cnabprovider.UninstallArguments{
		ActionArguments: cnabprovider.ActionArguments{
			Claim:                 o.Name,
			BundleIdentifier:      o.CNABFile,
			BundleIsFile:          true,
			Insecure:              o.Insecure,
			Params:                o.combineParameters(),
			CredentialIdentifiers: o.CredentialIdentifiers,
			Driver:                o.Driver,
		},
	}
}
