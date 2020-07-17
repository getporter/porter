package porter

import (
	"fmt"

	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
)

// UninstallOptions that may be specified when uninstalling a bundle.
// Porter handles defaulting any missing values.
type UninstallOptions struct {
	BundleLifecycleOpts
}

// UninstallBundle accepts a set of pre-validated UninstallOptions and uses
// them to uninstall a bundle.
func (p *Porter) UninstallBundle(opts UninstallOptions) error {
	err := p.prepullBundleByTag(&opts.BundleLifecycleOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before uninstall")
	}

	err = p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	deperator := newDependencyExecutioner(p, claim.ActionUninstall)
	err = deperator.Prepare(opts.BundleLifecycleOpts)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "uninstalling %s...\n", opts.Name)
	err = p.CNAB.Execute(opts.ToActionArgs(deperator))
	if err != nil {
		if len(deperator.deps) > 0 {
			return errors.Wrapf(err, "failed to uninstall the %s bundle, the remaining dependencies were not uninstalled", opts.Name)
		} else {
			return err
		}
	}

	// TODO: See https://github.com/deislabs/porter/issues/465 for flag to allow keeping around the dependencies
	return deperator.Execute()
}
