package porter

import (
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/pkg/errors"
)

var _ BundleAction = NewInstallOptions()

// InstallOptions that may be specified when installing a bundle.
// Porter handles defaulting any missing values.
type InstallOptions struct {
	*BundleActionOptions

	// Labels to apply to the installation.
	Labels []string
}

func (o InstallOptions) Validate(args []string, p *Porter) error {
	err := o.BundleActionOptions.Validate(args, p)
	if err != nil {
		return err
	}

	// Install requires special logic because the bundle must always be specified, including a name isn't enough.
	// So we have a slight repeat of the logic performed in by the generic bundle action args
	if o.File == "" && o.CNABFile == "" && o.Reference == "" {
		return errors.New("No bundle specified. Either --reference, --file or --cnab-file must be specified or the current directory must contain a porter.yaml file.")
	}

	return nil
}

func (o InstallOptions) ParseLabels() map[string]string {
	return parseLabels(o.Labels)
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
	i, err := p.Claims.GetInstallation(opts.Namespace, opts.Name)
	if err == nil {
		// Validate that we are not overwriting an existing installation
		if i.Status.InstallationCompleted {
			return errors.New("The installation has already been successfully installed and as a protection against accidentally overwriting existing installations, porter install cannot be repeated. Verify the installation name and namespace, and if correct, use porter upgrade.")
		}
	} else if errors.Is(err, storage.ErrNotFound{}) {
		// Create the installation record
		i = claims.NewInstallation(opts.Namespace, opts.Name)
		i.Labels = opts.ParseLabels()
		err = p.Claims.InsertInstallation(i)
		if err != nil {
			return errors.Wrap(err, "error saving installation record")
		}
	}

	return p.ExecuteAction(opts)
}
