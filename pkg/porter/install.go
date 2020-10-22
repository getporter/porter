package porter

import (
	"github.com/cnabio/cnab-go/claim"
)

var _ BundleAction = InstallOptions{}

// InstallOptions that may be specified when installing a bundle.
// Porter handles defaulting any missing values.
type InstallOptions struct {
	BundleActionOptions
}

func (o InstallOptions) GetAction() string {
	return claim.ActionInstall
}

func (o InstallOptions) GetActionVerb() string {
	return "installing"
}

// InstallBundle accepts a set of pre-validated InstallOptions and uses
// them to install a bundle.
func (p *Porter) InstallBundle(opts InstallOptions) error {
	return p.ExecuteAction(opts)
}
