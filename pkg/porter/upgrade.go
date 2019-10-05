package porter

import (
	"fmt"

	"github.com/deislabs/porter/pkg/manifest"

	"github.com/pkg/errors"
)

// UpgradeOptions that may be specified when upgrading a bundle.
// Porter handles defaulting any missing values.
type UpgradeOptions struct {
	BundleLifecycleOpts
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

	deperator := newDependencyExecutioner(p)
	err = deperator.Prepare(opts.BundleLifecycleOpts, p.CNAB.Upgrade)
	if err != nil {
		return err
	}

	err = deperator.Execute(manifest.ActionUpgrade)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "upgrading %s...\n", opts.Name)
	return p.CNAB.Upgrade(opts.ToActionArgs(deperator))
}
