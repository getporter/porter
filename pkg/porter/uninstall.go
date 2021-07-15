package porter

import (
	"fmt"
	"io"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

var _ BundleAction = NewUninstallOptions()

// ErrUnsafeInstallationDeleteRetryForceDelete presents the ErrUnsafeInstallationDelete error and provides a retry option of --force-delete
var ErrUnsafeInstallationDeleteRetryForceDelete = fmt.Errorf("%s; if you are sure it should be deleted, retry the last command with the --force-delete flag", ErrUnsafeInstallationDelete)

// UninstallOptions that may be specified when uninstalling a bundle.
// Porter handles defaulting any missing values.
type UninstallOptions struct {
	*BundleActionOptions
	UninstallDeleteOptions
}

func NewUninstallOptions() UninstallOptions {
	return UninstallOptions{BundleActionOptions: &BundleActionOptions{}}
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
func (p *Porter) UninstallBundle(opts UninstallOptions) error {
	err := p.prepullBundleByReference(opts.BundleActionOptions)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before uninstall")
	}

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	deperator := newDependencyExecutioner(p, cnab.ActionUninstall)
	err = deperator.Prepare(opts)
	if err != nil {
		return err
	}

	actionArgs, err := p.BuildActionArgs(opts)
	if err != nil {
		return err
	}
	deperator.PrepareRootActionArguments(&actionArgs)

	fmt.Fprintf(p.Out, "%s %s...\n", opts.GetActionVerb(), opts.Name)
	err = p.CNAB.Execute(actionArgs)

	var uninstallErrs error
	if err != nil {
		uninstallErrs = multierror.Append(uninstallErrs, err)

		// If the installation is not found, no further action is needed
		err := errors.Cause(err)
		if errors.Is(err, storage.ErrNotFound{}) {
			// TODO(carolynvs): find and fix all checks for not found
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
	err = opts.handleUninstallErrs(p.Out, deperator.Execute())
	if err != nil {
		return err
	}

	if opts.shouldDelete() {
		fmt.Fprintf(p.Out, installationDeleteTmpl, opts.Name)
		return p.Claims.RemoveInstallation(opts.Namespace, opts.Name)
	}
	return nil
}
