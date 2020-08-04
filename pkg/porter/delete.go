package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/context"
	claims "github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
)

const installationDeleteTmpl = "deleting installation records for %s...\n"

// ErrUnsafeInstallationDelete warns the user that deletion of an unsuccessfully uninstalled installation is unsafe
var ErrUnsafeInstallationDelete = errors.New("it is unsafe to delete an installation when the last action wasn't a successful uninstall; if you are sure it should be deleted, retry the last command with the --force flag")

// DeleteOptions represent options for Porter's installation delete command
type DeleteOptions struct {
	sharedOptions
	Force bool
}

// Validate prepares for an installation delete action and validates the args/options.
func (o *DeleteOptions) Validate(args []string, cxt *context.Context) error {
	// Ensure only one argument exists (installation name) if args length non-zero
	err := o.sharedOptions.validateInstallationName(args)
	if err != nil {
		return err
	}

	return o.sharedOptions.defaultBundleFiles(cxt)
}

// DeleteInstallation handles deletion of an installation
func (p *Porter) DeleteInstallation(opts DeleteOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	// TODO: do we check to see if the installation claim has required deps
	// declared and if so, attempt to delete them as well?

	claim, err := p.Claims.ReadLastClaim(opts.Name)
	if err != nil {
		return errors.Wrapf(err, "unable to read last claim for installation %s", opts.Name)
	}

	result, err := p.Claims.ReadLastResult(claim.ID)

	if (claim.Action != claims.ActionUninstall || result.Status != claims.StatusSucceeded) && !opts.Force {
		return ErrUnsafeInstallationDelete
	}

	fmt.Fprintf(p.Out, installationDeleteTmpl, opts.Name)
	return p.Claims.DeleteInstallation(opts.Name)
}
