package porter

import (
	"fmt"

	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

// UpgradeOptions that may be specified when upgrading a bundle.
// Porter handles defaulting any missing values.
type UpgradeOptions struct {
	BundleLifecycleOpts
}

func (o *UpgradeOptions) Validate(args []string, cxt *context.Context) error {
	if o.Tag != "" {
		err := o.validateTag()
		if err != nil {
			return err
		}
	}
	return o.sharedOptions.Validate(args, cxt)
}

// UpgradeBundle accepts a set of pre-validated UpgradeOptions and uses
// them to upgrade a bundle.
func (p *Porter) UpgradeBundle(opts UpgradeOptions) error {
	err := p.prepullBundleByTag(&opts.BundleLifecycleOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before upgrade")
	}

	err = p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "upgrading %s...\n", opts.Name)
	return p.CNAB.Upgrade(opts.ToDuffleArgs())
}

// ToDuffleArgs converts this instance of user-provided options
// to duffle arguments.
func (o *UpgradeOptions) ToDuffleArgs() cnabprovider.ActionArguments {
	return cnabprovider.ActionArguments{
		Claim:                 o.Name,
		BundleIdentifier:      o.CNABFile,
		BundleIsFile:          true,
		Insecure:              o.Insecure,
		Params:                o.combineParameters(),
		CredentialIdentifiers: o.CredentialIdentifiers,
		Driver:                o.Driver,
	}
}
