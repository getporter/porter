package porter

import (
	"fmt"

	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
)

var _ BundleAction = UpgradeOptions{}

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

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	deperator := newDependencyExecutioner(p, claim.ActionUpgrade)
	err = deperator.Prepare(opts)
	if err != nil {
		return err
	}

	err = deperator.Execute()
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "upgrading %s...\n", opts.Name)
	return p.CNAB.Execute(opts.ToActionArgs(deperator))
}
