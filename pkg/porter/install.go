package porter

import "get.porter.sh/porter/pkg/cnab"

var _ BundleAction = NewInstallOptions()

// InstallOptions that may be specified when installing a bundle.
// Porter handles defaulting any missing values.
type InstallOptions struct {
	*BundleActionOptions
}

func (o InstallOptions) GetAction() string {
	return cnab.ActionInstall
}

func (o InstallOptions) GetActionVerb() string {
	return "installing"
}

func NewInstallOptions() InstallOptions {
	return InstallOptions{BundleActionOptions: &BundleActionOptions{}}
}

// InstallBundle accepts a set of pre-validated InstallOptions and uses
// them to install a bundle.
func (p *Porter) InstallBundle(opts InstallOptions) error {
	return p.ExecuteAction(opts)
}
