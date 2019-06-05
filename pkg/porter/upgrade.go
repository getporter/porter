package porter

import (
	"fmt"

	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/context"
)

// UpgradeOptions that may be specified when uninstalling a bundle.
// Porter handles defaulting any missing values.
type UpgradeOptions struct {
	sharedOptions
}

func (o *UpgradeOptions) Validate(args []string, cxt *context.Context) error {
	o.bundleRequired = false
	return o.sharedOptions.Validate(args, cxt)
}

// UpgradeBundle accepts a set of pre-validated UpgradeOptions and uses
// them to upgrade a bundle.
func (p *Porter) UpgradeBundle(opts UpgradeOptions) error {
	p.applyDefaultOptions(&opts.sharedOptions)

	fmt.Fprintf(p.Out, "upgrading %s...\n", opts.Name)
	return p.CNAB.Upgrade(opts.ToDuffleArgs())
}

// ToDuffleArgs converts this instance of user-provided options
// to duffle arguments.
func (o *UpgradeOptions) ToDuffleArgs() cnabprovider.UpgradeArguments {
	return cnabprovider.UpgradeArguments{
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
