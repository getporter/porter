package porter

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/portercontext"
)

const installationDeleteTmpl = "deleting installation records for %s...\n"

var (
	// ErrUnsafeInstallationDelete warns the user that deletion of an unsuccessfully uninstalled installation is unsafe
	ErrUnsafeInstallationDelete = errors.New("it is unsafe to delete an installation when the last action wasn't a successful uninstall")

	// ErrUnsafeInstallationDeleteRetryForce presents the ErrUnsafeInstallationDelete error and provides a retry option of --force
	ErrUnsafeInstallationDeleteRetryForce = fmt.Errorf("%s; if you are sure it should be deleted, retry the last command with the --force flag", ErrUnsafeInstallationDelete)
)

// DeleteOptions represent options for Porter's installation delete command
type DeleteOptions struct {
	installationOptions
	Force bool
}

// Validate prepares for an installation delete action and validates the args/options.
func (o *DeleteOptions) Validate(args []string, cxt *portercontext.Context) error {
	// Ensure only one argument exists (installation name) if args length non-zero
	err := o.installationOptions.validateInstallationName(args)
	if err != nil {
		return err
	}

	return o.installationOptions.defaultBundleFiles(cxt)
}

// DeleteInstallation handles deletion of an installation
func (p *Porter) DeleteInstallation(ctx context.Context, opts DeleteOptions) error {
	err := p.applyDefaultOptions(ctx, &opts.installationOptions)
	if err != nil {
		return err
	}

	installation, err := p.Installations.GetInstallation(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return fmt.Errorf("unable to read status for installation %s: %w", opts.Name, err)
	}

	if (installation.Status.Action != cnab.ActionUninstall || installation.Status.ResultStatus != cnab.StatusSucceeded) && !opts.Force {
		return ErrUnsafeInstallationDeleteRetryForce
	}

	fmt.Fprintf(p.Out, installationDeleteTmpl, opts.Name)
	return p.Installations.RemoveInstallation(ctx, opts.Namespace, opts.Name)
}
