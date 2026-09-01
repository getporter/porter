package porter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/runtime"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-multierror"
)

// sharingGroupLabel is the installation label used to record the sharing
// group a dependency installation was created for, so a later run can tell
// whether it's still safe to reuse. Set on dependency installations created
// by runDependencyv2 below, and read by GraphBuilder's existing-installation
// matching (dependency_installation_resolver.go).
const sharingGroupLabel = "sh.porter.SharingGroup"

type dependencyExecutioner struct {
	*config.Config
	porter *Porter

	Resolver      BundleResolver
	CNAB          cnabprovider.CNABProvider
	Installations storage.InstallationProvider

	parentInstallation storage.Installation
	parentAction       BundleAction
	parentOpts         *BundleExecutionOptions

	// These are populated by Prepare, call it or perish in inevitable errors
	parentArgs cnabprovider.ActionArguments
	deps       []*queuedDependency

	// this should maybe go somewhere else
	depArgs cnabprovider.ActionArguments
}

func newDependencyExecutioner(p *Porter, installation storage.Installation, action BundleAction) *dependencyExecutioner {
	resolver := BundleResolver{
		Cache:    p.Cache,
		Registry: p.Registry,
	}
	return &dependencyExecutioner{
		porter:             p,
		parentInstallation: installation,
		parentAction:       action,
		parentOpts:         action.GetOptions(),
		Config:             p.Config,
		Resolver:           resolver,
		CNAB:               p.CNAB,
		Installations:      p.Installations,
	}
}

type queuedDependency struct {
	cnab.DependencyLock
	BundleReference cnab.BundleReference
	Parameters      map[string]string

	// cache of the CNAB file contents
	cnabFileContents []byte
}

func (e *dependencyExecutioner) Prepare(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	parentActionArgs, err := e.porter.BuildActionArgs(ctx, e.parentInstallation, e.parentAction)
	if err != nil {
		return err
	}
	e.parentArgs = parentActionArgs

	err = e.identifyDependencies(ctx)
	if err != nil {
		return err
	}

	for _, dep := range e.deps {
		err := e.prepareDependency(ctx, dep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *dependencyExecutioner) Execute(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if e.deps == nil {
		return span.Error(errors.New("Prepare must be called before Execute"))
	}

	// executeDependency the requested action against all the dependencies
	for _, dep := range e.deps {
		if !e.sharedActionResolver(ctx, dep) {
			// sharedActionResolver found a v2 dependency whose own bundle
			// action must not run here (install: it's already installed
			// for this SharingGroup; uninstall: this installation never
			// owns its lifecycle) -- but the parent's use of it still
			// needs to be reflected in the dependency's reference list,
			// which runDependencyv2 (never reached below) would otherwise
			// have handled.
			if dep.SharingMode {
				if err := e.recordSharedDependencyReference(ctx, dep); err != nil {
					return err
				}
			}
			return nil
		}
		err := e.executeDependency(ctx, dep)
		if err != nil {
			return err
		}
	}

	return nil
}

// PrepareRootActionArguments uses information about the dependencies of a bundle to prepare
// the execution of the root operation.
func (e *dependencyExecutioner) PrepareRootActionArguments(ctx context.Context) (cnabprovider.ActionArguments, error) {
	args, err := e.porter.BuildActionArgs(ctx, e.parentInstallation, e.parentAction)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	if args.Files == nil {
		args.Files = make(map[string]string, 2*len(e.deps))
	}

	// Define files necessary for dependencies that need to be copied into the bundle
	// args.Files is a map of target path to file contents
	// This creates what goes in /cnab/app/dependencies/DEP.NAME
	for _, dep := range e.deps {
		// Copy the dependency bundle.json
		err = e.checkSharedOutputs(ctx, dep)
		if err != nil {
			return cnabprovider.ActionArguments{}, err
		}
		target := runtime.GetDependencyDefinitionPath(dep.Alias)
		args.Files[target] = string(dep.cnabFileContents)
	}
	return args, nil
}

func (e *dependencyExecutioner) checkSharedOutputs(ctx context.Context, dep *queuedDependency) error {
	if !e.sharedActionResolver(ctx, dep) && e.parentAction.GetAction() == "install" {
		return e.getActionArgs(ctx, dep)
	}
	return nil
}

// sharedActionResolver tries to localize if v2, and shared deps
// then what actions should we take based off labels/action type/state
// true means continue, false means stop
func (e *dependencyExecutioner) sharedActionResolver(ctx context.Context, dep *queuedDependency) bool {
	depInstallation, err := e.Installations.GetInstallation(ctx, e.parentOpts.Namespace, dep.Alias)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound{}) {
			return true
		}
	}
	e.depArgs.Installation = depInstallation

	//We're real, let's check if this is in the installation the parent
	// is referencing
	if dep.SharingGroup == depInstallation.Labels[sharingGroupLabel] {
		if e.parentAction.GetAction() == "install" {
			return false
		}
		if e.parentAction.GetAction() == "upgrade" {
			return true
		}
		if e.parentAction.GetAction() == "uninstall" {
			return false
		}
	}
	return true
}

// recordSharedDependencyReference adds or drops this parent's reference to
// a v2 (SharingMode) dependency for the install/uninstall cases where
// sharedActionResolver has decided the dependency's own bundle action must
// not run (it's already installed for this SharingGroup, or this
// installation never owns its lifecycle). Must be called with
// e.depArgs.Installation already set to the dependency found by
// sharedActionResolver.
func (e *dependencyExecutioner) recordSharedDependencyReference(ctx context.Context, dep *queuedDependency) error {
	depInstallation := e.depArgs.Installation

	if e.parentAction.GetAction() == cnab.ActionUninstall {
		if depInstallation.RemoveReference(e.parentArgs.Installation.String()) {
			return e.Installations.UpdateInstallation(ctx, depInstallation)
		}
		return nil
	}

	if depInstallation.AddReference(e.parentArgs.Installation.String(), dep.Alias) {
		return e.Installations.UpdateInstallation(ctx, depInstallation)
	}
	return nil
}

func (e *dependencyExecutioner) identifyDependencies(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// Load parent CNAB bundle definition
	var bun cnab.ExtendedBundle
	if e.parentOpts.CNABFile != "" {
		bundle, err := e.CNAB.LoadBundle(e.parentOpts.CNABFile)
		if err != nil {
			return span.Error(err)
		}
		bun = bundle
	} else if e.parentOpts.Reference != "" {
		cachedBundle, err := e.Resolver.Resolve(ctx, e.parentOpts.BundlePullOptions)
		if err != nil {
			return span.Error(fmt.Errorf("could not resolve bundle: %w", err))
		}

		bun = cachedBundle.Definition

	} else if e.parentOpts.Name != "" {
		c, err := e.Installations.GetLastRun(ctx, e.parentOpts.Namespace, e.parentOpts.Name)
		if err != nil {
			return err
		}

		bun = cnab.NewBundle(c.Bundle)
	} else {
		// If we hit here, there is a bug somewhere
		return span.Error(errors.New("identifyDependencies failed to load the bundle because no bundle was specified. Please report this bug to https://github.com/getporter/porter/issues/new/choose"))
	}

	// V2 dependency wiring (a parameter/credential/output sourced from a
	// sibling dependency's output) is only checked for resolvability today
	// during porter inspect/explain, where it's a non-fatal warning so the
	// dependency tree can still be viewed. Here, before any dependency or the
	// root bundle actually runs, an unresolvable reference must be a hard
	// failure -- a bundle can't run correctly with broken wiring. v1 bundles
	// have no wiring concept, so this is skipped entirely for them.
	if bun.HasDependenciesV2() {
		if err := e.porter.validateDependencyWiring(ctx, bun, e.parentOpts); err != nil {
			return span.Error(err)
		}
	}

	// Inject registry provider for dependency resolution.
	// registryListTagsAdapter bridges the cnabtooci.RegistryProvider
	// (concrete opts) and the registryListTags interface in the cnab
	// package (opts interface{}) to avoid a circular import.
	regOpts := cnabtooci.RegistryOptions{InsecureRegistry: e.parentOpts.InsecureRegistry}
	adapter := &registryListTagsAdapter{reg: e.Resolver.Registry, opts: regOpts}
	eb := bun.WithRegistry(adapter, regOpts)

	// Determine version strategy: flag overrides global config
	strategy := e.parentOpts.DependenciesVersionStrategy
	if strategy == "" {
		strategy = e.GetDependenciesVersionStrategy()
	}
	eb = eb.WithVersionStrategy(strategy)

	locks, err := eb.ResolveDependencies(ctx, bun)
	if err != nil {
		return span.Error(err)
	}

	e.deps = make([]*queuedDependency, len(locks))
	for i, lock := range locks {
		span.Debugf("Resolved dependency %s to %s", lock.Alias, lock.Reference)
		e.deps[i] = &queuedDependency{
			DependencyLock: lock,
		}
	}

	return nil
}

func (e *dependencyExecutioner) prepareDependency(ctx context.Context, dep *queuedDependency) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()
	// Pull the dependency
	var err error
	pullOpts := BundlePullOptions{
		Reference:        dep.Reference,
		InsecureRegistry: e.parentOpts.InsecureRegistry,
		Force:            e.parentOpts.Force,
	}
	if err := pullOpts.Validate(); err != nil {
		return span.Error(fmt.Errorf("error preparing dependency %s: %w", dep.Alias, err))
	}
	cachedDep, err := e.Resolver.Resolve(ctx, pullOpts)
	if err != nil {
		return span.Error(fmt.Errorf("error pulling dependency %s: %w", dep.Alias, err))
	}
	dep.BundleReference = cachedDep.BundleReference

	strategy := e.GetSchemaCheckStrategy(ctx)
	err = cachedDep.Definition.Validate(e.Context, strategy)
	if err != nil {
		return span.Error(fmt.Errorf("invalid bundle %s: %w", dep.Alias, err))
	}

	// Cache the bundle.json for later
	dep.cnabFileContents, err = e.FileSystem.ReadFile(cachedDep.BundlePath)
	if err != nil {
		return span.Error(fmt.Errorf("error reading %s: %w", cachedDep.BundlePath, err))
	}

	// Make a lookup of which parameters are defined in the dependent bundle
	depParams := map[string]struct{}{}
	for paramName := range cachedDep.Definition.Parameters {
		depParams[paramName] = struct{}{}
	}

	// Handle any parameter overrides for the dependency defined in porter.yaml
	// dependencies:
	//  requires:
	//   - name: DEP
	//     parameters:
	//       PARAM: VALUE
	// TODO: When we redo dependencies, we need to remove this dependency on the bundle being a porter bundle with a manifest
	// Yes, right now the way this works means this feature is Porter only
	m := &manifest.Manifest{}
	if e.parentOpts.File != "" {
		var err error
		m, err = manifest.LoadManifestFrom(ctx, e.Config, e.parentOpts.File)
		if err != nil {
			return err
		}
	}

	for _, manifestDep := range m.Dependencies.Requires {
		if manifestDep.Name == dep.Alias {
			for paramName, value := range manifestDep.Parameters {
				// Make sure the parameter is defined in the bundle
				if _, ok := depParams[paramName]; !ok {
					return fmt.Errorf("invalid dependencies.%s.parameters entry, %s is not a parameter defined in that bundle", dep.Alias, paramName)
				}

				if dep.Parameters == nil {
					dep.Parameters = make(map[string]string, 1)
				}
				dep.Parameters[paramName] = value
			}
		}
	}

	// Handle any parameter overrides for the dependency defined on the command line
	// --param DEP#PARAM=VALUE
	for key, value := range e.parentOpts.depParams {
		parts := strings.Split(key, "#")
		if len(parts) > 1 && parts[0] == dep.Alias {
			paramName := parts[1]

			// Make sure the parameter is defined in the bundle
			if _, ok := depParams[paramName]; !ok {
				return fmt.Errorf("invalid --param %s, %s is not a parameter defined in the bundle %s", key, paramName, dep.Alias)
			}

			if dep.Parameters == nil {
				dep.Parameters = make(map[string]string, 1)
			}
			dep.Parameters[paramName] = value
		}
	}

	return nil
}

func (e *dependencyExecutioner) executeDependency(ctx context.Context, dep *queuedDependency) error {
	// TODO(carolynvs): We should really switch up how the deperator works so that
	// even the root bundle uses the execution engine here. This would set up how
	// we want dependencies and mixins as bundles to work in the future.

	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if dep.SharingMode {
		err := e.runDependencyv2(ctx, dep)
		return err
	}

	eb := cnab.ExtendedBundle{}
	//this expects depv1 style dependency to be installed as parentName+depName
	depName := eb.BuildPrerequisiteInstallationName(e.parentOpts.Name, dep.Alias)
	depInstallation, err := e.Installations.GetInstallation(ctx, e.parentOpts.Namespace, depName)

	if err != nil {
		if errors.Is(err, storage.ErrNotFound{}) {
			depInstallation = storage.NewInstallation(e.parentOpts.Namespace, depName)
			depInstallation.SetLabel("sh.porter.parentInstallation", e.parentArgs.Installation.String())

			// For now, assume it's okay to give the dependency the same credentials as the parent
			depInstallation.CredentialSets = e.parentInstallation.CredentialSets
			if err = e.Installations.InsertInstallation(ctx, depInstallation); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	e.depArgs.Installation = depInstallation

	if err = e.getActionArgs(ctx, dep); err != nil {
		return err
	}

	if err = e.finalizeExecute(ctx, dep); err != nil {
		return err
	}

	return nil
}

// runDependencyv2 will see if the child dependency is already installed
// and if so, use sharingmode && group to resolve what to do
func (e *dependencyExecutioner) runDependencyv2(ctx context.Context, dep *queuedDependency) error {
	action := e.parentAction.GetAction()

	depInstallation, err := e.Installations.GetInstallation(ctx, e.parentOpts.Namespace, dep.Alias)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound{}) {
			// Nothing to reference-drop or uninstall if the dependency was
			// never created to begin with.
			if action == cnab.ActionUninstall {
				return nil
			}

			depInstallation = storage.NewInstallation(e.parentOpts.Namespace, dep.Alias)
			depInstallation.SetLabel("sh.porter.parentInstallation", e.parentArgs.Installation.String())
			depInstallation.SetLabel(sharingGroupLabel, dep.SharingGroup)
			depInstallation.AddReference(e.parentArgs.Installation.String(), dep.Alias)

			// For now, assume it's okay to give the dependency the same credentials as the parent
			depInstallation.CredentialSets = e.parentInstallation.CredentialSets
			if err = e.Installations.InsertInstallation(ctx, depInstallation); err != nil {
				return err
			}

			// err is nil here (the Insert above succeeded), so this returns
			// without ever calling getActionArgs/finalizeExecute below --
			// meaning a newly-created v2 dependency's own install action
			// never actually runs via this path. Pre-existing behavior,
			// predates reference tracking; not fixed here to keep this
			// change scoped to #2608.
			return err
		}
	}

	// This installation being uninstalled never owns a shared dependency's
	// lifecycle, whether it originally created it or is just referencing one
	// that already existed -- it only ever borrows it. So uninstalling never
	// runs the dependency's own uninstall action or deletes its installation
	// record here; it only drops this installation's reference. Once nothing
	// references it, actually deleting the shared installation (and, first,
	// uninstalling it) is left to an explicit, separate action (porter
	// installations delete / porter uninstall on the dependency itself).
	// err == nil here means depInstallation was actually found by
	// GetInstallation above (the not-found case already returned); guarding
	// on it avoids acting on a zero-value depInstallation if GetInstallation
	// failed with some other error, a pre-existing gap in this function
	// unrelated to reference tracking.
	if err == nil && action == cnab.ActionUninstall {
		if depInstallation.RemoveReference(e.parentArgs.Installation.String()) {
			return e.Installations.UpdateInstallation(ctx, depInstallation)
		}
		return nil
	}

	// Record that the parent depends on this installation to satisfy
	// dep.Alias, whether it was already installed independently or created
	// by a prior run.
	if err == nil && depInstallation.AddReference(e.parentArgs.Installation.String(), dep.Alias) {
		if err := e.Installations.UpdateInstallation(ctx, depInstallation); err != nil {
			return err
		}
	}
	//We save the installation
	e.depArgs.Installation = depInstallation

	// Installed: Return
	// Uninstalled: Error (delete or else)
	// Upgrade: Unsupported
	// Invoke: At your own risk
	//todo(schristoff): this is kind of icky, can be it less so?
	if dep.SharingGroup == depInstallation.Labels[sharingGroupLabel] {
		if depInstallation.IsInstalled() {
			if action == "upgrade" {
				return nil
			}
		}
		if depInstallation.Uninstalled {
			return fmt.Errorf("error executing dependency, dependency must be in installed status or deleted, %s is in  status %s", dep.Alias, depInstallation.Status.ResultStatus)
		}

	}

	if err = e.getActionArgs(ctx, dep); err != nil {
		return err
	}

	if err = e.finalizeExecute(ctx, dep); err != nil {
		return err
	}

	return nil
}

func (e *dependencyExecutioner) getActionArgs(ctx context.Context,
	dep *queuedDependency) error {
	actionName := e.parentArgs.Run.Action
	finalParams, err := e.porter.finalizeParameters(ctx, e.depArgs.Installation, dep.BundleReference.Definition, actionName, dep.Parameters)
	if err != nil {
		return fmt.Errorf("error resolving parameters for dependency %s: %w", dep.Alias, err)
	}
	depRun, err := e.porter.createRun(ctx, dep.BundleReference, e.depArgs.Installation, actionName, finalParams)
	if err != nil {
		return fmt.Errorf("error creating run for dependency %s: %w", dep.Alias, err)
	}
	e.depArgs = cnabprovider.ActionArguments{
		BundleReference:       dep.BundleReference,
		Installation:          e.depArgs.Installation,
		Run:                   depRun,
		Driver:                e.parentArgs.Driver,
		AllowDockerHostAccess: e.parentOpts.AllowDockerHostAccess,
		PersistLogs:           e.parentArgs.PersistLogs,
	}
	return nil
}

// finalizeExecute handles some Uninstall logic that is carried out
// right before calling CNAB execute.
func (e *dependencyExecutioner) finalizeExecute(ctx context.Context, dep *queuedDependency) error {
	ctx, span := tracing.StartSpan(ctx)
	// Determine if we're working with UninstallOptions, to inform deletion and
	// error handling, etc.
	var uninstallOpts UninstallOptions
	if opts, ok := e.parentAction.(UninstallOptions); ok {
		uninstallOpts = opts
	}

	var executeErrs error
	span.Infof("Executing dependency %s...", dep.Alias)
	err := e.CNAB.Execute(ctx, e.depArgs)
	if err != nil {
		executeErrs = multierror.Append(executeErrs, fmt.Errorf("error executing dependency %s: %w", dep.Alias, err))

		// Handle errors when/if the action is uninstall
		// If uninstallOpts is an empty struct, executeErrs will pass through
		executeErrs = uninstallOpts.handleUninstallErrs(e.Err, executeErrs)
		if executeErrs != nil {
			return span.Error(executeErrs)
		}
	}

	// If uninstallOpts is an empty struct (i.e., action not Uninstall), this
	// will resolve to false and thus be a no-op.
	//
	// Only v1 dependencies reach this: runDependencyv2 handles the v2
	// (SharingMode) uninstall case itself, before ever calling
	// getActionArgs/finalizeExecute, since a v2 dependency's lifecycle is
	// never owned by an installation that merely references it (see
	// runDependencyv2). v1 dependency names are scoped to this parent
	// (BuildPrerequisiteInstallationName) and can never be shared, so
	// deleting one unconditionally here is always safe.
	if uninstallOpts.shouldDelete() {
		span.Infof(installationDeleteTmpl, e.depArgs.Installation)
		return e.Installations.RemoveInstallation(ctx, e.depArgs.Installation.Namespace, e.depArgs.Installation.Name)
	}
	return nil
}

// registryListTagsAdapter bridges cnabtooci.RegistryProvider (concrete
// RegistryOptions parameter) and the cnab.registryListTags interface
// (opts interface{}) so that real registries satisfy the interface used
// by ExtendedBundle.determineDefaultTag without creating a circular import.
type registryListTagsAdapter struct {
	reg  cnabtooci.RegistryProvider
	opts cnabtooci.RegistryOptions
}

func (a *registryListTagsAdapter) ListTags(ctx context.Context, ref cnab.OCIReference, _ interface{}) ([]string, error) {
	return a.reg.ListTags(ctx, ref, a.opts)
}
