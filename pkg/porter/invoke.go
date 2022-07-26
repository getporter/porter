package porter

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/storage"
)

var _ BundleAction = NewInvokeOptions()

// InvokeOptions that may be specified when invoking a bundle.
// Porter handles defaulting any missing values.
type InvokeOptions struct {
	*BundleExecutionOptions

	// Action name to invoke
	Action string
}

func NewInvokeOptions() InvokeOptions {
	return InvokeOptions{
		BundleExecutionOptions: NewBundleExecutionOptions(),
	}
}

func (o InvokeOptions) GetAction() string {
	return o.Action
}

func (o InvokeOptions) GetActionVerb() string {
	return "invoking"
}

func (o InvokeOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	if o.Action == "" {
		return errors.New("--action is required")
	}

	return o.BundleExecutionOptions.Validate(ctx, args, p)
}

// InvokeBundle accepts a set of pre-validated InvokeOptions and uses
// them to upgrade a bundle.
func (p *Porter) InvokeBundle(ctx context.Context, opts InvokeOptions) error {
	// Figure out which bundle/installation we are working with
	bundleRef, err := p.resolveBundleReference(ctx, opts.BundleReferenceOptions)
	if err != nil {
		return err
	}

	installation, err := p.Installations.GetInstallation(ctx, opts.Namespace, opts.Name)
	if errors.Is(err, storage.ErrNotFound{}) {
		action, actionErr := bundleRef.Definition.GetAction(opts.Action)
		if actionErr != nil {
			return fmt.Errorf("invalid --action %s", opts.Action)
		}

		// Only allow actions on a non-existing installation when it won't change anything
		if action.Modifies || !action.Stateless {
			return fmt.Errorf("could not find installation %s/%s: %w", opts.Namespace, opts.Name, err)
		}

		// Create an ephemeral installation just for this run
		installation = storage.Installation{Namespace: opts.Namespace, Name: opts.Name}
	}
	err = p.applyActionOptionsToInstallation(ctx, &installation, opts.BundleExecutionOptions)
	if err != nil {
		return err
	}
	return p.ExecuteAction(ctx, installation, opts)
}
