package porter

import (
	"fmt"

	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
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

	return o.BundleLifecycleOpts.Validate(args, cxt)
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

	deperator := newDependencyExecutioner(p)
	err = deperator.Prepare(opts.BundleLifecycleOpts, func(args cnabprovider.ActionArguments) error {
		return p.CNAB.Invoke(opts.Action, args)
	})
	if err != nil {
		return err
	}

	err = deperator.Execute(manifest.Action(opts.Action))
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "invoking custom action %s on %s...\n", opts.Action, opts.Name)
	return p.CNAB.Invoke(opts.Action, opts.ToActionArgs(deperator))
}
