package porter

import (
	"fmt"
	"strings"

	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

// InvokeOptions that may be specified when invoking a bundle.
// Porter handles defaulting any missing values.
type InvokeOptions struct {
	// Action name to invoke
	Action string
	BundleLifecycleOpts
}

func (o *InvokeOptions) Validate(args []string, cxt *context.Context) error {
	if o.Action == "" {
		return errors.New("--action is required")
	}

	o.Action = strings.ToLower(o.Action)

	if o.Tag != "" {
		err := o.validateTag()
		if err != nil {
			return err
		}
	}
	return o.sharedOptions.Validate(args, cxt)
}

// InvokeBundle accepts a set of pre-validated InvokeOptions and uses
// them to upgrade a bundle.
func (p *Porter) InvokeBundle(opts InvokeOptions) error {
	err := p.prepullBundleByTag(&opts.BundleLifecycleOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before invoking the custom action")
	}

	err = p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "invoking custom action %s on %s...\n", opts.Action, opts.Name)
	return p.CNAB.Invoke(opts.Action, opts.ToDuffleArgs())
}

// ToDuffleArgs converts this instance of user-provided options
// to duffle arguments.
func (o *InvokeOptions) ToDuffleArgs() cnabprovider.ActionArguments {
	return cnabprovider.ActionArguments{
		Claim:                 o.Name,
		BundleIdentifier:      o.CNABFile,
		BundleIsFile:          true,
		Insecure:              o.Insecure,
		Params:                o.combineParameters(),
		CredentialIdentifiers: o.CredentialIdentifiers,
		Driver:                o.Driver,
	}
}
