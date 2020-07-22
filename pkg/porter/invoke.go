package porter

import (
	"fmt"

	"github.com/pkg/errors"
)

// InvokeOptions that may be specified when invoking a bundle.
// Porter handles defaulting any missing values.
type InvokeOptions struct {
	// Action name to invoke
	Action string
	BundleLifecycleOpts
}

func (o *InvokeOptions) Validate(args []string, p *Porter) error {
	if o.Action == "" {
		return errors.New("--action is required")
	}

	return o.BundleLifecycleOpts.Validate(args, p)
}

// InvokeBundle accepts a set of pre-validated InvokeOptions and uses
// them to upgrade a bundle.
func (p *Porter) InvokeBundle(opts InvokeOptions) error {
	err := p.prepullBundleByTag(&opts.BundleLifecycleOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before invoking the custom action")
	}

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	deperator := newDependencyExecutioner(p, opts.Action)
	err = deperator.Prepare(opts.BundleLifecycleOpts)
	if err != nil {
		return err
	}

	err = deperator.Execute()
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "invoking custom action %s on %s...\n", opts.Action, opts.Name)
	return p.CNAB.Execute(opts.ToActionArgs(deperator))
}
