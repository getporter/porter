package porter

import (
	"fmt"
	"strings"

	"github.com/cnabio/cnab-go/claim"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// ErrUnsafeInstallationDeleteRetryForceDelete presents the ErrUnsafeInstallationDelete error and provides a retry option of --force-delete
var ErrUnsafeInstallationDeleteRetryForceDelete = fmt.Errorf("%s; if you are sure it should be deleted, retry the last command with the --force-delete flag", ErrUnsafeInstallationDelete)

// UninstallOptions that may be specified when uninstalling a bundle.
// Porter handles defaulting any missing values.
type UninstallOptions struct {
	BundleLifecycleOpts
}

// UninstallDeleteOptions supply options for deletion on uninstall
type UninstallDeleteOptions struct {
	Delete      bool
	ForceDelete bool
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
	err = deperator.Prepare(opts.BundleLifecycleOpts)
	if err != nil {
		return err
	}

	var uninstallErrs error
	fmt.Fprintf(p.Out, "uninstalling %s...\n", opts.Name)
	err = p.CNAB.Execute(opts.ToActionArgs(deperator))
	if err != nil {
		uninstallErrs = multierror.Append(uninstallErrs, err)
		// If the installation is not found, bail out now and return this error; no further action needed
		if strings.Contains(err.Error(), claim.ErrInstallationNotFound.Error()) {
			return uninstallErrs
		}

		if len(deperator.deps) > 0 {
			uninstallErrs = multierror.Append(uninstallErrs,
				fmt.Errorf("failed to uninstall the %s bundle, the remaining dependencies were not uninstalled", opts.Name))
		}

		if opts.Delete && !opts.ForceDelete {
			uninstallErrs = multierror.Append(uninstallErrs, ErrUnsafeInstallationDeleteRetryForceDelete)
		}
		if !opts.ForceDelete {
			return uninstallErrs
		}
		// else, we swallow uninstallErrs as opts.ForceDelete is true
		// and we wish to pass same option down to deps, if applicable
	}

	// TODO: See https://github.com/deislabs/porter/issues/465 for flag to allow keeping around the dependencies
	err = deperator.Execute()
	if err != nil {
		if !opts.ForceDelete {
			return err
		}
	}

	if opts.Delete || opts.ForceDelete {
		fmt.Fprintf(p.Out, installationDeleteTmpl, opts.Name)
		return p.Claims.DeleteInstallation(opts.Name)
	}
	return nil
}
