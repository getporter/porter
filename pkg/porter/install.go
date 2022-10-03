package porter

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
)

var _ BundleAction = NewInstallOptions()

// InstallOptions that may be specified when installing a bundle.
// Porter handles defaulting any missing values.
type InstallOptions struct {
	*BundleExecutionOptions

	// Labels to apply to the installation.
	Labels []string
}

func (o InstallOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	err := o.BundleExecutionOptions.Validate(ctx, args, p)
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
	return InstallOptions{
		BundleExecutionOptions: NewBundleExecutionOptions(),
	}
}

// InstallBundle accepts a set of pre-validated InstallOptions and uses
// them to install a bundle.
func (p *Porter) InstallBundle(ctx context.Context, opts InstallOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// Figure out which bundle/installation we are working with
	bundleRef, err := p.resolveBundleReference(ctx, opts.BundleReferenceOptions)
	if err != nil {
		return log.Error(err)
	}

	i, err := p.Installations.GetInstallation(ctx, opts.Namespace, opts.Name)
	if err == nil {
		// Validate that we are not overwriting an existing installation
		if i.IsInstalled() && !opts.Force {
			err = errors.New("The installation has already been successfully installed and as a protection against accidentally overwriting existing installations, porter install cannot be repeated. Verify the installation name and namespace, and if correct, use porter upgrade. You can skip this check by using the --force flag.")
			return log.Error(err)
		}
	} else if errors.Is(err, storage.ErrNotFound{}) {
		// Create the installation record
		i = storage.NewInstallation(opts.Namespace, opts.Name)
	} else {
		err = fmt.Errorf("could not retrieve the installation record: %w", err)
		return log.Error(err)
	}

	err = p.applyActionOptionsToInstallation(ctx, &i, opts.BundleExecutionOptions)
	if err != nil {
		return err
	}
	i.TrackBundle(bundleRef.Reference)
	i.Labels = opts.ParseLabels()
	err = p.Installations.UpsertInstallation(ctx, i)
	if err != nil {
		return fmt.Errorf("error saving installation record: %w", err)
	}

	// Run install using the updated installation record
	return p.ExecuteAction(ctx, i, opts)
}

// Remember the parameters and credentials used with the bundle last.
// Appends any newly specified parameters, parameter/credential sets to the installation record.
// Users are expected to edit the installation record if they don't want that behavior.
func (p *Porter) applyActionOptionsToInstallation(ctx context.Context, i *storage.Installation, opts *BundleExecutionOptions) error {
	// Record the parameters specified by the user, with flags taking precedence over parameter set values
	err := opts.LoadParameters(ctx, p, opts.bundleRef.Definition)
	if err != nil {
		return err
	}
	// Record the user-specified parameter values
	err = opts.populateInternalParameterSet(ctx, p, opts.bundleRef.Definition, i)
	if err != nil {
		return err
	}

	// Record the names of the parameter and credential sets used if specified. Otherwise, reuse the previously specified sets.
	// This should replace previously specified sets so that only what was just specified is used.
	if len(opts.ParameterSets) > 0 {
		i.ParameterSets = opts.ParameterSets
	}
	if len(opts.CredentialIdentifiers) > 0 {
		i.CredentialSets = opts.CredentialIdentifiers
	}

	return nil
}
