package porter

import (
	"get.porter.sh/porter/pkg/cnab"
	"github.com/pkg/errors"
)

var _ BundleAction = NewUpgradeOptions()

// UpgradeOptions that may be specified when upgrading a bundle.
// Porter handles defaulting any missing values.
type UpgradeOptions struct {
	*BundleActionOptions
}

func NewUpgradeOptions() UpgradeOptions {
	return UpgradeOptions{&BundleActionOptions{}}
}

func (o UpgradeOptions) GetAction() string {
	return cnab.ActionUpgrade
}

func (o UpgradeOptions) GetActionVerb() string {
	return "upgrading"
}

// UpgradeBundle accepts a set of pre-validated UpgradeOptions and uses
// them to upgrade a bundle.
func (p *Porter) UpgradeBundle(opts UpgradeOptions) error {
	installation, err := p.Claims.GetInstallation(opts.Namespace, opts.Name)
	if err != nil {
		return errors.Wrapf(err, "could not find installation %s/%s", opts.Namespace, opts.Name)
	}
	return p.ExecuteAction(installation, opts)
}
