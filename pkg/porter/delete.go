package porter

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
)

const installationDeleteTmpl = "deleting installation records for %s...\n"

var (
	// ErrUnsafeInstallationDelete warns the user that deletion of an unsuccessfully uninstalled installation is unsafe
	ErrUnsafeInstallationDelete = errors.New("it is unsafe to delete an installation when the last action wasn't a successful uninstall")

	// ErrUnsafeInstallationDeleteRetryForce presents the ErrUnsafeInstallationDelete error and provides a retry option of --force
	ErrUnsafeInstallationDeleteRetryForce = fmt.Errorf("%s; if you are sure it should be deleted, retry the last command with the --force flag", ErrUnsafeInstallationDelete)

	// ErrInstallationReferenced warns the user that deletion of an installation that other installations still depend on is unsafe
	ErrInstallationReferenced = errors.New("it is unsafe to delete an installation that other installations still depend on")

	// ErrInstallationReferencedRetryForce presents the ErrInstallationReferenced error and provides a retry option of --force
	ErrInstallationReferencedRetryForce = fmt.Errorf("%s; if you are sure it should be deleted, retry the last command with the --force flag", ErrInstallationReferenced)

	// ErrInstallationReferencedRetryForceDelete presents the ErrInstallationReferenced error and provides a retry option of --force-delete
	ErrInstallationReferencedRetryForceDelete = fmt.Errorf("%s; if you are sure it should be deleted, retry the last command with the --force-delete flag", ErrInstallationReferenced)
)

// DeleteOptions represent options for Porter's installation delete command
type DeleteOptions struct {
	installationOptions
	Force bool
}

// Validate prepares for an installation delete action and validates the args/options.
func (o *DeleteOptions) Validate(args []string, cxt *portercontext.Context) error {
	// Ensure only one argument exists (installation name) if args length non-zero
	err := o.validateInstallationName(args)
	if err != nil {
		return err
	}

	return o.defaultBundleFiles(cxt)
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

	if installation.IsReferenced() && !opts.Force {
		return ErrInstallationReferencedRetryForce
	}

	fmt.Fprintf(p.Out, installationDeleteTmpl, opts.Name)
	if err := p.Installations.RemoveInstallation(ctx, opts.Namespace, opts.Name); err != nil {
		return err
	}

	// installation may itself have been referencing other installations to
	// satisfy its own dependencies. Normally that's cleaned up by
	// dependencyExecutioner.runDependencyv2 during `porter uninstall`, but
	// this standalone delete command runs after the fact and has no record
	// of what installation depended on, so sweep for any installation still
	// referencing it. Best-effort: the installation record is already gone,
	// there's nothing to roll back to, so failures here are reported as
	// warnings rather than failing the delete.
	referenced, err := p.Installations.FindInstallations(ctx, storage.FindOptions{
		Filter: bson.M{"status.references.installation": installation.String()},
	})
	if err != nil {
		fmt.Fprintf(p.Err, "warning: unable to check for stale references to the deleted installation %s: %s\n", installation, err)
		return nil
	}
	for _, ref := range referenced {
		if ref.RemoveReference(installation.String()) {
			if err := p.Installations.UpdateInstallation(ctx, ref); err != nil {
				fmt.Fprintf(p.Err, "warning: unable to remove the stale reference to the deleted installation %s from %s: %s\n", installation, ref, err)
			}
		}
	}
	return nil
}
