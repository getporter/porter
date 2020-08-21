package porter

import (
	"fmt"
	"io"
	"strings"

	"github.com/cnabio/cnab-go/claim"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

var _ BundleAction = UninstallOptions{}

// ErrUnsafeInstallationDeleteRetryForceDelete presents the ErrUnsafeInstallationDelete error and provides a retry option of --force-delete
var ErrUnsafeInstallationDeleteRetryForceDelete = fmt.Errorf("%s; if you are sure it should be deleted, retry the last command with the --force-delete flag", ErrUnsafeInstallationDelete)

// UninstallOptions that may be specified when uninstalling a bundle.
// Porter handles defaulting any missing values.
type UninstallOptions struct {
	BundleLifecycleOpts
	UninstallDeleteOptions
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

func (opts UninstallOptions) GetBundleLifecycleOptions() BundleLifecycleOpts {
	return opts.BundleLifecycleOpts
}

// UninstallBundle accepts a set of pre-validated UninstallOptions and uses
// them to uninstall a bundle.
func (p *Porter) UninstallBundle(opts UninstallOptions) error {
	err := p.prepullBundleByTag(&opts.BundleLifecycleOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before uninstall")
	}

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	deperator := newDependencyExecutioner(p, claim.ActionUninstall)
	err = deperator.Prepare(opts)
	if err != nil {
		return err
	}

	var uninstallErrs error
	fmt.Fprintf(p.Out, "uninstalling %s...\n", opts.Name)
	err = p.CNAB.Execute(opts.ToActionArgs(deperator))
	if err != nil {
		uninstallErrs = multierror.Append(uninstallErrs, err)

		// If the installation is not found, no further action is needed
		if strings.Contains(err.Error(), claim.ErrInstallationNotFound.Error()) {
			return uninstallErrs
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

	// TODO: See https://github.com/deislabs/porter/issues/465 for flag to allow keeping around the dependencies
	err = opts.handleUninstallErrs(p.Out, deperator.Execute())
	if err != nil {
		return err
	}

	if opts.shouldDelete() {
		fmt.Fprintf(p.Out, installationDeleteTmpl, opts.Name)
		return p.Claims.DeleteInstallation(opts.Name)
	}
	return nil
}
