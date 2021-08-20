package porter

import (
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

func (o InvokeOptions) Validate(args []string, p *Porter) error {
	if o.Action == "" {
		return errors.New("--action is required")
	}

	return o.BundleActionOptions.Validate(args, p)
}

// InvokeBundle accepts a set of pre-validated InvokeOptions and uses
// them to upgrade a bundle.
func (p *Porter) InvokeBundle(opts InvokeOptions) error {
	bundleRef, err := p.resolveBundleReference(opts.BundleActionOptions)
	if err != nil {
		return err
	}

	action, err := bundleRef.Definition.GetAction(opts.Action)
	if err != nil {
		return fmt.Errorf("invalid --action %s", opts.Action)
	}

	installation, err := p.Claims.GetInstallation(opts.Namespace, opts.Name)
	if errors.Is(err, storage.ErrNotFound{}) {
		if action.Modifies || !action.Stateless {
			return errors.Wrapf(err, "could not find installation %s/%s", opts.Namespace, opts.Name)
		}

		// Create an ephemeral installation just for this run
		installation = claims.Installation{Namespace: opts.Namespace, Name: opts.Name}
	}
	return p.ExecuteAction(installation, opts)
}
