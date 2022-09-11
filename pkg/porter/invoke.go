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
	bundleRef, err := opts.GetBundleReference(ctx, p)
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
		installation = storage.NewInstallation(opts.Namespace, opts.Name)
	}

	err = p.applyActionOptionsToInstallation(ctx, opts, &installation)
	if err != nil {
		return err
	}

	if p.useWorkflowEngine(opts.bundleRef.Definition) {
		puller := NewBundleResolver(p.Cache, opts.Force, p.Registry, opts.GetRegistryOptions())
		eng := NewWorkflowEngine(installation.Namespace, puller, p.Installations, p)
		workflowOpts := CreateWorkflowOptions{
			Installation: installation,
			CustomAction: opts.Action,
			Bundle:       opts.bundleRef.Definition,
			DebugMode:    opts.DebugMode,
			MaxParallel:  1, // TODO(PEP003): make this configurable
		}
		ws, err := eng.CreateWorkflow(ctx, workflowOpts)
		if err != nil {
			return err
		}

		w := storage.Workflow{WorkflowSpec: ws}
		if err := p.Installations.InsertWorkflow(ctx, w); err != nil {
			return err
		}

		// TODO(PEP003): if a dry-run is requested, print out the execution plan and then exit
		return eng.RunWorkflow(ctx, w)
	}

	return p.ExecuteBundleAndDependencies(ctx, installation, opts)
}
