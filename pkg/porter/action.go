package porter

import (
	"bytes"
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"

	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/tracing"

	"get.porter.sh/porter/pkg/storage"
)

// ExecuteBundleAndDependencies runs a specified action for a root bundle.
// The bundle should be a root bundle, and if there are dependencies, they will also be executed as appropriate.
// Supported actions are: install, upgrade, invoke.
// The uninstall action works in reverse, so it's implemented separately.
// Dependencies are resolved and executed differently depending on whether the deps-v2 feature is enabled (workflow).
func (p *Porter) ExecuteBundleAndDependencies(ctx context.Context, installation storage.Installation, action BundleAction) error {
	// Callers should check for a noop action (because the installation is up-to-date, but let's check too just in case
	if action == nil {
		return nil
	}

	opts := action.GetOptions()
	bundleRef := opts.bundleRef

	ctx, span := tracing.StartSpan(ctx,
		tracing.ObjectAttribute("installation", installation),
		attribute.String("action", action.GetAction()),
		attribute.Bool("dry-run", opts.DryRun),
	)
	defer span.EndSpan()

	// Switch between our two dependency implementations
	depsv2 := p.useWorkflowEngine(bundleRef.Definition)
	span.SetAttributes(attribute.Bool("deps-v2", depsv2))

	if depsv2 {
		// TODO(PEP003): Use new getregistryoptions elsewhere that we create that
		puller := NewBundleResolver(p.Cache, opts.Force, p.Registry, opts.GetRegistryOptions())
		eng := NewWorkflowEngine(installation.Namespace, puller, p.Installations, p)
		workflowOpts := CreateWorkflowOptions{
			Installation: installation,
			Bundle:       bundleRef.Definition,
			DebugMode:    opts.DebugMode,
			MaxParallel:  1,
		}
		ws, err := eng.CreateWorkflow(ctx, workflowOpts)
		if err != nil {
			return err
		}

		if opts.DryRun {
			span.Info("Skipping workflow execution because --dry-run was specified")

			// TODO(PEP003): It would be better to have a way to always emit something to stdout, and capture it in the trace at the same time
			var buf bytes.Buffer
			err = printer.PrintYaml(&buf, ws)
			fmt.Fprintln(p.Out, buf.String())
			span.SetAttributes(attribute.String("workflow", buf.String()))

			// TODO(PEP003): Print out the generated workflow according to opts.Format
			// TODO(PEP003): how do we want to get Format in here so we can print properly?
			return err
		}

		w := storage.Workflow{WorkflowSpec: ws}
		if err := p.Installations.InsertWorkflow(ctx, w); err != nil {
			return err
		}

		return eng.RunWorkflow(ctx, w)
	} else { // Fallback to the old implementation of dependencies and bundle execution
		if opts.DryRun {
			span.Info("Skipping bundle execution because --dry-run was specified")
			return nil
		}

		deperator := newDependencyExecutioner(p, installation, action)
		err := deperator.Prepare(ctx)
		if err != nil {
			return err
		}

		err = deperator.Execute(ctx)
		if err != nil {
			return err
		}

		actionArgs, err := deperator.PrepareRootActionArguments(ctx)
		if err != nil {
			return err
		}

		return p.CNAB.Execute(ctx, actionArgs)
	}
}

// ExecuteRootBundleOnly runs a single bundle that has already had its dependencies resolved by a workflow.
// The workflow is responsible identifying the bundles to run, their order, what to pass between them, etc.
// It is only intended to be used with the deps-v2 feature.
func (p *Porter) ExecuteRootBundleOnly(ctx context.Context, installation storage.Installation, action BundleAction) error {
	// Callers should check for a noop action (because the installation is up-to-date, but let's check too just in case
	if action == nil {
		return nil
	}

	opts := action.GetOptions()
	ctx, span := tracing.StartSpan(ctx,
		tracing.ObjectAttribute("installation", installation),
		attribute.String("action", action.GetAction()),
		attribute.Bool("dry-run", opts.DryRun),
	)
	defer span.EndSpan()

	actionArgs, err := p.BuildActionArgs(ctx, installation, action)
	if err != nil {
		return err
	}

	return p.CNAB.Execute(ctx, actionArgs)
}
