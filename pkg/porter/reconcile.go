package porter

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg/printer"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/otel/attribute"
)

type ReconcileOptions struct {
	Installation storage.InstallationSpec

	// Just reapply the installation regardless of what has changed (or not)
	Force bool

	// DryRun only checks if the changes would trigger a bundle run
	DryRun bool

	// Format that should be used when printing details about what Porter is (or will) do.
	Format printer.Format
}

// ReconcileInstallationAndDependencies compares the desired state of an installation
// as stored in the installation record with the current state of the
// installation. If they are not in sync, the appropriate bundle action
// is executed to bring them in sync.
// This is only used for install/upgrade actions triggered by applying a file
// to an installation. For uninstall or invoke, you should call those directly.
func (p *Porter) ReconcileInstallationAndDependencies(ctx context.Context, opts ReconcileOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	installation, actionOpts, err := p.reconcileInstallation(ctx, opts)
	if err != nil {
		return err
	}

	// Nothing to do, the installation is up-to-date
	if actionOpts == nil {
		return nil
	}

	return p.ExecuteBundleAndDependencies(ctx, installation, actionOpts)
}

// ReconcileInstallationInWorkflow compares the desired state of an installation
// as stored in the installation record with the current state of the
// installation. If they are not in sync, the appropriate bundle action
// is executed to bring them in sync.
// This is only used for install/upgrade actions triggered by applying a file
// to an installation. For uninstall or invoke, you should call those directly.
// This should only be used with deps-v2 feature workflows.
func (p *Porter) ReconcileInstallationInWorkflow(ctx context.Context, opts ReconcileOptions) (storage.Run, storage.Result, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	installation, actionOpts, err := p.reconcileInstallation(ctx, opts)
	if err != nil {
		return storage.Run{}, storage.Result{}, err
	}

	// Nothing to do, the installation is up-to-date
	if actionOpts == nil {
		return storage.Run{}, storage.Result{}, nil
	}

	return p.ExecuteRootBundleOnly(ctx, installation, actionOpts)
}

func (p *Porter) reconcileInstallation(ctx context.Context, opts ReconcileOptions) (storage.Installation, BundleAction, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// Determine if the installation exists
	inputInstallation := opts.Installation
	span.Debugf("Reconciling %s/%s installation", inputInstallation.Namespace, inputInstallation.Name)
	installation, err := p.Installations.GetInstallation(ctx, inputInstallation.Namespace, inputInstallation.Name)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound{}) {
			return storage.Installation{}, nil, fmt.Errorf("could not query for an existing installation document for %s: %w", inputInstallation, err)
		}

		// Create a new installation
		installation = storage.NewInstallation(inputInstallation.Namespace, inputInstallation.Name)
		installation.Apply(inputInstallation)

		span.Info("Creating a new installation", attribute.String("installation", installation.String()))
	} else {
		// Apply the specified changes to the installation
		installation.Apply(inputInstallation)
		if err := installation.Validate(); err != nil {
			return storage.Installation{}, nil, err
		}

		fmt.Fprintf(p.Err, "Updating %s installation\n", installation)
	}

	// Get the last run of the installation, if available
	var lastRun *storage.Run
	r, err := p.Installations.GetLastRun(ctx, inputInstallation.Namespace, inputInstallation.Name)
	neverRun := errors.Is(err, storage.ErrNotFound{})
	if err != nil && !neverRun {
		return storage.Installation{}, nil, err
	}
	if !neverRun {
		lastRun = &r
	}

	ref, ok, err := opts.Installation.Bundle.GetBundleReference()
	if err != nil {
		return storage.Installation{}, nil, span.Error(err)
	}
	if !ok {
		instYaml, _ := yaml.Marshal(opts.Installation)
		return storage.Installation{}, nil, span.Error(fmt.Errorf("the installation does not define a valid bundle reference.\n%s", instYaml))
	}

	// Configure the bundle action that we should execute IF IT'S OUT OF SYNC
	var actionOpts BundleAction
	if installation.IsInstalled() {
		if opts.Installation.Uninstalled {
			actionOpts = NewUninstallOptions()
		} else {
			actionOpts = NewUpgradeOptions()
		}
	} else {
		actionOpts = NewInstallOptions()
	}

	lifecycleOpts := actionOpts.GetOptions()
	lifecycleOpts.DryRun = opts.DryRun
	lifecycleOpts.Reference = ref.String()
	lifecycleOpts.Name = inputInstallation.Name
	lifecycleOpts.Namespace = inputInstallation.Namespace
	lifecycleOpts.Driver = p.Data.RuntimeDriver
	lifecycleOpts.CredentialIdentifiers = opts.Installation.CredentialSets
	lifecycleOpts.ParameterSets = opts.Installation.ParameterSets

	if err = p.applyActionOptionsToInstallation(ctx, actionOpts, &installation); err != nil {
		return storage.Installation{}, nil, err
	}

	// Determine if the installation's desired state is out of sync with reality 🤯
	inSync, err := p.IsInstallationInSync(ctx, installation, lastRun, actionOpts)
	if err != nil {
		return storage.Installation{}, nil, err
	}

	if inSync {
		if opts.Force {
			span.Info("The installation is up-to-date but will be re-applied because --force was specified")
		} else {
			span.Info("The installation is already up-to-date.")
			return storage.Installation{}, nil, nil
		}
	}

	span.Infof("The installation is out-of-sync, running the %s action...", actionOpts.GetAction())
	if err := actionOpts.Validate(ctx, nil, p); err != nil {
		return storage.Installation{}, nil, err
	}

	if opts.DryRun {
		span.Info("Skipping bundle execution because --dry-run was specified")
		return storage.Installation{}, nil, nil
	} else {
		if err = p.Installations.UpsertInstallation(ctx, installation); err != nil {
			return storage.Installation{}, nil, err
		}
	}

	return installation, actionOpts, nil
}

// IsInstallationInSync determines if the desired state of the installation matches
// the state of the installation the last time it was modified.
func (p *Porter) IsInstallationInSync(ctx context.Context, i storage.Installation, lastRun *storage.Run, action BundleAction) (bool, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// Only print out info messages if we are triggering a bundle run. Otherwise, keep the explanations in debug output.

	// Has it been uninstalled? If so, we don't ever reconcile it again
	if i.IsUninstalled() {
		log.Info("Ignoring because the installation is uninstalled")
		return true, nil
	}

	// Should we uninstall it?
	if i.Uninstalled {
		// Only try to uninstall if it's been installed before
		if i.IsInstalled() {
			log.Info("Triggering because installation.uninstalled is true")
			return false, nil
		}

		// Otherwise ignore this installation
		log.Info("Ignoring because installation.uninstalled is true but the installation doesn't exist yet")
		return true, nil
	} else {
		// TODO(PEP003): we should check the status of the last run and handle in progress/pending by returning an error if the not in sync otherise
		// i.e. if we run two commands to apply, the first starts, the second succeeds since it asked for what the other is providing?
		// apply waits, so really it should wait for the pending/inprogress to complete? or stop early and say it's in progress elsewhere.

		// Should we install it?
		if !i.IsInstalled() {
			log.Info("Triggering because the installation has not completed successfully yet")
			return false, nil
		}
	}

	// We want to upgrade but we don't have values to compare against
	// This shouldn't happen but check just in case
	if lastRun == nil {
		log.Info("Triggering because the last run for the installation wasn't recorded")
		return false, nil
	}

	// Figure out if we need to upgrade
	opts := action.GetOptions()

	newRef, err := opts.GetBundleReference(ctx, p)
	if err != nil {
		return false, err
	}

	// Has the bundle definition changed?
	if lastRun.BundleDigest != newRef.Digest.String() {
		log.Info("Triggering because the bundle definition has changed",
			attribute.String("oldReference", lastRun.BundleReference),
			attribute.String("oldDigest", lastRun.BundleDigest),
			attribute.String("newReference", newRef.Reference.String()),
			attribute.String("newDigest", newRef.Digest.String()))
		return false, nil
	}

	// Convert parameters to a string to compare them. This avoids problems comparing
	// values that may be equal but have different types due to how the parameter
	// value was loaded.
	b := newRef.Definition
	prepParametersForComparison := func(params map[string]interface{}) (map[string]string, error) {
		compParams := make(map[string]string, len(params))
		for paramName, rawValue := range params {
			if b.IsInternalParameter(paramName) {
				continue
			}

			typedValue, err := b.ConvertParameterValue(paramName, rawValue)
			if err != nil {
				return nil, err
			}

			stringValue, err := b.WriteParameterToString(paramName, typedValue)
			if err != nil {
				return nil, err
			}

			compParams[paramName] = stringValue
		}
		return compParams, nil
	}

	lastRunParams, err := p.Sanitizer.RestoreParameterSet(ctx, lastRun.Parameters, cnab.NewBundle(lastRun.Bundle))
	if err != nil {
		return false, err
	}

	oldParams, err := prepParametersForComparison(lastRunParams)
	if err != nil {
		return false, fmt.Errorf("error prepping old parameters for comparision: %w", err)
	}

	newParams, err := prepParametersForComparison(opts.GetParameters())
	if err != nil {
		return false, fmt.Errorf("error prepping current parameters for comparision: %w", err)
	}

	if !cmp.Equal(oldParams, newParams) {
		diff := cmp.Diff(oldParams, newParams)
		log.Info("Triggering because the parameters have changed",
			attribute.String("diff", diff))
		return false, nil
	}

	// Check only if the names of the associated credential sets have changed
	// This is a "good enough for now" decision that can be revisited if we
	// get use cases for needing to diff the actual credentials.
	sort.Strings(lastRun.CredentialSets)
	sort.Strings(i.CredentialSets)
	if !cmp.Equal(lastRun.CredentialSets, i.CredentialSets) {
		diff := cmp.Diff(lastRun.CredentialSets, i.CredentialSets)
		log.Info("Triggering because the credential set names have changed",
			attribute.String("diff", diff))
		return false, nil
	}
	return true, nil
}
