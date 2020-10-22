package porter

import (
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
)

var _ BundleAction = UpgradeOptions{}

// UpgradeOptions that may be specified when upgrading a bundle.
// Porter handles defaulting any missing values.
type UpgradeOptions struct {
	BundleActionOptions
}

func (o UpgradeOptions) GetAction() string {
	return claim.ActionUpgrade
}

func (o UpgradeOptions) GetActionVerb() string {
	return "upgrading"
}

// UpgradeBundle accepts a set of pre-validated UpgradeOptions and uses
// them to upgrade a bundle.
func (p *Porter) UpgradeBundle(opts UpgradeOptions) error {
	err := p.prepullBundleByTag(&opts.BundleActionOptions)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before upgrade")
	}

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	return p.ExecuteAction(opts)
}
