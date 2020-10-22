package porter

import (
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
	return p.ExecuteAction(opts)
}
