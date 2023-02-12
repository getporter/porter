package porter

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
)

var _ BundleAction = NewInstallOptions()

// InstallOptions that may be specified when installing a bundle.
// Porter handles defaulting any missing values.
type InstallOptions struct {
	*BundleExecutionOptions

	// Labels to apply to the installation.
	Labels []string
}

func (o InstallOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	err := o.BundleExecutionOptions.Validate(ctx, args, p)
	if err != nil {
		return err
	}

	// Install requires special logic because the bundle must always be specified, including a name isn't enough.
	// So we have a slight repeat of the logic performed in by the generic bundle action args
	if o.File == "" && o.CNABFile == "" && o.Reference == "" {
		return errors.New("No bundle specified. Either --reference, --file or --cnab-file must be specified or the current directory must contain a porter.yaml file.")
	}

	return nil
}

func (o InstallOptions) ParseLabels() map[string]string {
	return parseLabels(o.Labels)
}

func (o InstallOptions) GetAction() string {
	return cnab.ActionInstall
}

func (o InstallOptions) GetActionVerb() string {
	return "installing"
}

func NewInstallOptions() InstallOptions {
	return InstallOptions{
		BundleExecutionOptions: NewBundleExecutionOptions(),
	}
}

// InstallBundle accepts a set of pre-validated InstallOptions and uses
// them to install a bundle.
func (p *Porter) InstallBundle(ctx context.Context, opts InstallOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	i, err := p.Installations.GetInstallation(ctx, opts.Namespace, opts.Name)
	if err == nil {
		// Validate that we are not overwriting an existing installation
		if i.IsInstalled() && !opts.Force {
			err = errors.New("The installation has already been successfully installed and as a protection against accidentally overwriting existing installations, porter install cannot be repeated. Verify the installation name and namespace, and if correct, use porter upgrade. You can skip this check by using the --force flag.")
			return log.Error(err)
		}
	} else if errors.Is(err, storage.ErrNotFound{}) {
		// Create the installation record
		i = storage.NewInstallation(opts.Namespace, opts.Name)
	} else {
		err = fmt.Errorf("could not retrieve the installation record: %w", err)
		return log.Error(err)
	}

	// Apply labels that were specified as flags to the installation record
	i.Labels = opts.ParseLabels()

	err = p.applyActionOptionsToInstallation(ctx, opts, &i)
	if err != nil {
		return err
	}

	bundleRef, err := opts.GetBundleReference(ctx, p)
	if err != nil {
		return err
	}

	if p.useWorkflowEngine(bundleRef.Definition) {
		// TODO(PEP003): Use new getregistryoptions elsewhere that we create that
		puller := NewBundleResolver(p.Cache, opts.Force, p.Registry, opts.GetRegistryOptions())
		eng := NewWorkflowEngine(i.Namespace, puller, p.Installations, p)
		workflowOpts := CreateWorkflowOptions{
			Installation: i,
			Bundle:       bundleRef.Definition,
			DebugMode:    opts.DebugMode,
			MaxParallel:  1,
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

	// Use the old implementation of bundle execution compatible with depsv1
	err = p.Installations.UpsertInstallation(ctx, i)
	if err != nil {
		return fmt.Errorf("error saving installation record: %w", err)
	}

	// Run install using the updated installation record
	return p.ExecuteBundleAndDependencies(ctx, i, opts)
}

// useWorkflowEngine determines if the new workflow engine or the old bundle execution code should be used.
// Once depsv2 is no longer experimental, we can switch 100% to the workflow engine
// Old bundles can still use depsv1, since depsv2 is a superset of depsv1.
// It will change how the bundle is run, for example calling install right now twice in a row
// results in an error, and this would remove that limitation, and instead a second call to install causes it to be reconciled and possibly skipped.
// In either case, the solution to the user is to call --force so the change isn't breaking.
func (p *Porter) useWorkflowEngine(bun cnab.ExtendedBundle) bool {
	if bun.HasDependenciesV2() {
		return true
	}

	return p.Config.IsFeatureEnabled(experimental.FlagDependenciesV2)
}

func (p *Porter) sanitizeInstallation(ctx context.Context, inst *storage.Installation, bun cnab.ExtendedBundle) error {
	strategies, err := p.Sanitizer.CleanParameters(ctx, inst.Parameters.Parameters, bun, inst.ID)
	if err != nil {
		return err
	}

	inst.Parameters = inst.NewInternalParameterSet(strategies...)
	return nil
}
