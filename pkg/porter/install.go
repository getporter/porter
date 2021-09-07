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
	// Figure out which bundle/installation we are working with
	bundleRef, err := p.resolveBundleReference(opts.BundleActionOptions)
	if err != nil {
		return err
	}

	i, err := p.Claims.GetInstallation(opts.Namespace, opts.Name)
	if err == nil {
		// Validate that we are not overwriting an existing installation
		if i.Status.InstallationCompleted && !opts.Force {
			return errors.New("The installation has already been successfully installed and as a protection against accidentally overwriting existing installations, porter install cannot be repeated. Verify the installation name and namespace, and if correct, use porter upgrade. You can skip this check by using the --force flag.")
		}
	} else if errors.Is(err, storage.ErrNotFound{}) {
		// Create the installation record
		i = claims.NewInstallation(opts.Namespace, opts.Name)
	} else {
		return errors.Wrapf(err, "could not retrieve the installation record")
	}

	err = p.applyActionOptionsToInstallation(i, opts.BundleActionOptions)
	if err != nil {
		return err
	}
	i.TrackBundle(bundleRef.Reference)
	i.Labels = opts.ParseLabels()
	err = p.Claims.UpsertInstallation(i)
	if err != nil {
		return errors.Wrap(err, "error saving installation record")
	}

	// Run install using the updated installation record
	return p.ExecuteAction(i, opts)
}

// Remember the parameters and credentials used with the bundle last.
// Appends any newly specified parameters, parameter/credential sets to the installation record.
// Users are expected to edit the installation record if they don't want that behavior.
func (p *Porter) applyActionOptionsToInstallation(i claims.Installation, opts *BundleActionOptions) error {
	// Record the parameters specified by the user, with flags taking precedence over parameter set values
	err := opts.LoadParameters(p)
	if err != nil {
		return err
	}
	if i.Parameters == nil {
		i.Parameters = make(map[string]interface{}, len(opts.parsedParams))
	}
	// Record the user-specified parameter values
	for k, v := range opts.parsedParams {
		i.Parameters[k] = v
	}
	// Record the names of the parameter sets used
	for _, ps := range opts.ParameterSets {
		if isParameterSetFile, _ := p.FileSystem.Exists(ps); isParameterSetFile {
			continue
		}
		for _, existing := range i.ParameterSets {
			if existing == ps {
				continue
			}
		}
		i.ParameterSets = append(i.ParameterSets, ps)
	}
	// Record the names of the credential sets used
	for _, cs := range opts.CredentialIdentifiers {
		if isCredentialSetFile, _ := p.FileSystem.Exists(cs); isCredentialSetFile {
			continue
		}
		for _, existing := range i.ParameterSets {
			if existing == cs {
				continue
			}
		}
		i.CredentialSets = append(i.CredentialSets, cs)
	}
	return nil
}
