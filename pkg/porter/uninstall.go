package porter

import (
	"context"
	"errors"
	"fmt"
	"io"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-multierror"
)

var _ BundleAction = NewUninstallOptions()

// ErrUnsafeInstallationDeleteRetryForceDelete presents the ErrUnsafeInstallationDelete error and provides a retry option of --force-delete
var ErrUnsafeInstallationDeleteRetryForceDelete = fmt.Errorf("%s; if you are sure it should be deleted, retry the last command with the --force-delete flag", ErrUnsafeInstallationDelete)

// UninstallOptions that may be specified when uninstalling a bundle.
// Porter handles defaulting any missing values.
type UninstallOptions struct {
	*BundleExecutionOptions
	UninstallDeleteOptions
}

func NewUninstallOptions() UninstallOptions {
	return UninstallOptions{
		BundleExecutionOptions: NewBundleExecutionOptions(),
	}
}

func (o UninstallOptions) GetAction() string {
	return cnab.ActionUninstall
}

func (o UninstallOptions) GetActionVerb() string {
	return "uninstalling"
}

// UninstallDeleteOptions supply options for deletion on uninstall
type UninstallDeleteOptions struct {
	Delete      bool
	ForceDelete bool
}

func (opts *UninstallDeleteOptions) shouldDelete() bool {
	return opts.Delete || opts.ForceDelete
}

func (opts *UninstallDeleteOptions) unsafeDelete() bool {
	return opts.Delete && !opts.ForceDelete
}

func (opts *UninstallDeleteOptions) handleUninstallErrs(out io.Writer, err error) error {
	if err == nil {
		return nil
	}

	if opts.unsafeDelete() {
		return multierror.Append(err, ErrUnsafeInstallationDeleteRetryForceDelete)
	}

	if opts.ForceDelete {
		fmt.Fprintf(out, "ignoring the following errors as --force-delete is true:\n  %s", err.Error())
		return nil
	}
	return err
}

// UninstallBundle accepts a set of pre-validated UninstallOptions and uses
// them to uninstall a bundle.
func (p *Porter) UninstallBundle(ctx context.Context, opts UninstallOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// Figure out which bundle/installation we are working with
	_, err := p.resolveBundleReference(ctx, opts.BundleReferenceOptions)
	if err != nil {
		return err
	}

	installation, err := p.Installations.GetInstallation(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return fmt.Errorf("could not find installation %s/%s: %w", opts.Namespace, opts.Name, err)
	}

	err = p.applyActionOptionsToInstallation(ctx, &installation, opts.BundleExecutionOptions)
	if err != nil {
		return err
	}

	deperator := newDependencyExecutioner(p, installation, opts)
	err = deperator.Prepare(ctx)
	if err != nil {
		return err
	}

	actionArgs, err := deperator.PrepareRootActionArguments(ctx)
	if err != nil {
		return err
	}

	log.Infof("%s bundle", opts.GetActionVerb())
	err = p.CNAB.Execute(ctx, actionArgs)

	var uninstallErrs error
	if err != nil {
		uninstallErrs = multierror.Append(uninstallErrs, err)

		// If the installation is not found, no further action is needed
		err := errors.Unwrap(err)
		if errors.Is(err, storage.ErrNotFound{}) {
			return err
		}

		if len(deperator.deps) > 0 && !opts.ForceDelete {
			uninstallErrs = multierror.Append(uninstallErrs,
				fmt.Errorf("failed to uninstall the %s bundle, the remaining dependencies were not uninstalled", opts.Name))
		}

		uninstallErrs = opts.handleUninstallErrs(p.Out, uninstallErrs)
		if uninstallErrs != nil {
			return uninstallErrs
		}
	}

	// TODO: See https://github.com/getporter/porter/issues/465 for flag to allow keeping around the dependencies
	err = opts.handleUninstallErrs(p.Out, deperator.Execute(ctx))
	if err != nil {
		return err
	}

	if opts.shouldDelete() {
		log.Info("deleting installation records")
		return p.Installations.RemoveInstallation(ctx, opts.Namespace, opts.Name)
	}
	return nil
}
