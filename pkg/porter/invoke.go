package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/storage"
	"github.com/pkg/errors"
)

var _ BundleAction = NewInvokeOptions()

// InvokeOptions that may be specified when invoking a bundle.
// Porter handles defaulting any missing values.
type InvokeOptions struct {
	// Action name to invoke
	Action string
	*BundleActionOptions
}

func NewInvokeOptions() InvokeOptions {
	return InvokeOptions{BundleActionOptions: &BundleActionOptions{}}
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

	return o.BundleActionOptions.Validate(ctx, args, p)
}

// InvokeBundle accepts a set of pre-validated InvokeOptions and uses
// them to upgrade a bundle.
func (p *Porter) InvokeBundle(ctx context.Context, opts InvokeOptions) error {
	// Figure out which bundle/installation we are working with
	bundleRef, err := p.resolveBundleReference(ctx, opts.BundleActionOptions)
	if err != nil {
		return err
	}

	installation, err := p.Claims.GetInstallation(ctx, opts.Namespace, opts.Name)
	if errors.Is(err, storage.ErrNotFound{}) {
		action, actionErr := bundleRef.Definition.GetAction(opts.Action)
		if actionErr != nil {
			return fmt.Errorf("invalid --action %s", opts.Action)
		}

		// Only allow actions on a non-existing installation when it won't change anything
		if action.Modifies || !action.Stateless {
			return errors.Wrapf(err, "could not find installation %s/%s", opts.Namespace, opts.Name)
		}

		// Create an ephemeral installation just for this run
		installation = claims.Installation{Namespace: opts.Namespace, Name: opts.Name}
	}
	err = p.applyActionOptionsToInstallation(ctx, &installation, opts.BundleActionOptions)
	if err != nil {
		return err
	}
	return p.ExecuteAction(ctx, installation, opts)
}
