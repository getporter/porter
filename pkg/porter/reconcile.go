package porter

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/otel/attribute"
)

type ReconcileOptions struct {
	Name         string
	Namespace    string
	Installation storage.Installation

	// Just reapply the installation regardless of what has changed (or not)
	Force bool

	// DryRun only checks if the changes would trigger a bundle run
	DryRun bool
}

// ReconcileInstallation compares the desired state of an installation
// as stored in the installation record with the current state of the
// installation. If they are not in sync, the appropriate bundle action
// is executed to bring them in sync.
// This is only used for install/upgrade actions triggered by applying a file
// to an installation. For uninstall or invoke, you should call those directly.
func (p *Porter) ReconcileInstallation(ctx context.Context, opts ReconcileOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	log.Debugf("Reconciling %s/%s installation", opts.Namespace, opts.Name)

	// Get the last run of the installation, if available
	var lastRun *storage.Run
	r, err := p.Installations.GetLastRun(ctx, opts.Namespace, opts.Name)
	neverRun := errors.Is(err, storage.ErrNotFound{})
	if err != nil && !neverRun {
		return err
	}
	if !neverRun {
		lastRun = &r
	}

	ref, ok, err := opts.Installation.Bundle.GetBundleReference()
	if err != nil {
		return log.Error(err)
	}
	if !ok {
		instYaml, _ := yaml.Marshal(opts.Installation)
		return log.Error(fmt.Errorf("the installation does not define a valid bundle reference.\n%s", instYaml))
	}

	// Configure the bundle action that we should execute IF IT'S OUT OF SYNC
	var actionOpts BundleAction
	if opts.Installation.IsInstalled() {
		if opts.Installation.Uninstalled {
			actionOpts = NewUninstallOptions()
		} else {
			actionOpts = NewUpgradeOptions()
		}
	} else {
		actionOpts = NewInstallOptions()
	}

	lifecycleOpts := actionOpts.GetOptions()
	lifecycleOpts.Reference = ref.String()
	lifecycleOpts.Name = opts.Name
	lifecycleOpts.Namespace = opts.Namespace
	lifecycleOpts.CredentialIdentifiers = opts.Installation.CredentialSets

	lifecycleOpts.ParameterSets = opts.Installation.ParameterSets
	lifecycleOpts.Params = make([]string, 0, len(opts.Installation.Parameters.Parameters))

	// Write out the parameters as string values. Not efficient but reusing ExecuteAction would need more refactoring otherwise
	_, err = p.resolveBundleReference(ctx, lifecycleOpts.BundleReferenceOptions)
	if err != nil {
		return err
	}

	for _, param := range opts.Installation.Parameters.Parameters {
		lifecycleOpts.Params = append(lifecycleOpts.Params, fmt.Sprintf("%s=%s", param.Name, param.Value))
	}

	if err := p.applyActionOptionsToInstallation(ctx, &opts.Installation, lifecycleOpts); err != nil {
		return err
	}

	if !opts.DryRun {
		if err = p.Installations.UpsertInstallation(ctx, opts.Installation); err != nil {
			return err
		}
	}

	// Determine if the installation's desired state is out of sync with reality 🤯
	inSync, err := p.IsInstallationInSync(ctx, opts.Installation, lastRun, actionOpts)
	if err != nil {
		return err
	}

	if inSync {
		if opts.Force {
			log.Info("The installation is up-to-date but will be re-applied because --force was specified")
		} else {
			log.Info("The installation is already up-to-date.")
			return nil
		}
	}

	log.Infof("The installation is out-of-sync, running the %s action...", actionOpts.GetAction())
	if err := actionOpts.Validate(ctx, nil, p); err != nil {
		return err
	}

	if opts.DryRun {
		log.Info("Skipping bundle execution because --dry-run was specified")
		return nil
	}

	return p.ExecuteAction(ctx, opts.Installation, actionOpts)
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

	newRef, err := p.resolveBundleReference(ctx, opts.BundleReferenceOptions)
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

	// Get a set of parameters ready for comparison to another set of parameters
	// to tell if the installation should be executed again. For now I'm just
	// removing internal parameters (e.g. porter-debug, porter-state) and making
	// sure that the types are correct, etc.
	b := newRef.Definition
	resolvedParams, err := p.resolveParameters(ctx, i, b, action.GetAction(), opts.combinedParameters)
	if err != nil {
		return false, err
	}

	// Convert parameters to a string to compare them. This avoids problems comparing
	// values that may be equal but have different types due to how the parameter
	// value was loaded.
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

	newParams, err := prepParametersForComparison(resolvedParams)
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
