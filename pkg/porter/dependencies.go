package porter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/runtime"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-multierror"
)

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
		target := runtime.GetDependencyDefinitionPath(dep.DependencyLock.Alias)
		args.Files[target] = string(dep.cnabFileContents)
	}

	return args, nil
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
	locks, err := bun.ResolveDependencies(bun)
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
			delete(e.parentArgs.Params, key)
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

	if err = e.getActionArgs(ctx, dep, depInstallation); err != nil {
		return err
	}

	if err = e.finalizeExecute(ctx, dep); err != nil {
		return err
	}

	return nil
}

// runDependencyv2 will see if the child dependency is already installed
// if so, it will not install the dependency, but rather share the new parent
// to the child
func (e *dependencyExecutioner) runDependencyv2(ctx context.Context, dep *queuedDependency) error {
	depInstallation, err := e.Installations.GetInstallation(ctx, e.parentOpts.Namespace, dep.Alias)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound{}) {
			depInstallation = storage.NewInstallation(e.parentOpts.Namespace, dep.Alias)
			depInstallation.SetLabel("sh.porter.parentInstallation", e.parentArgs.Installation.String())
			depInstallation.SetLabel("sh.porter.SharingGroup", dep.SharingGroup)

			// For now, assume it's okay to give the dependency the same credentials as the parent
			depInstallation.CredentialSets = e.parentInstallation.CredentialSets
			if err = e.Installations.InsertInstallation(ctx, depInstallation); err != nil {
				return err
			}

			return err
		}
	}

	if dep.SharingGroup == depInstallation.Labels["sh.porter.SharingGroup"] {
		//todo(schristoff): should we just make a new one if uninstalled?
		// but then what should it's name be.
		if !depInstallation.Uninstalled {
			return fmt.Errorf("error executing dependency, dependency must be in installed status or deleted, %s is in  status %s", dep.Alias, depInstallation.Status)
		}
	}

	//todo(schristoff): PEP003 Find a better place to put this.
	// if dep.SharingMode && e.parentAction.GetAction()
	if err = e.getActionArgs(ctx, dep, depInstallation); err != nil {
		return err
	}

	if err = e.finalizeExecute(ctx, dep); err != nil {
		return err
	}

	return nil
}

func (e *dependencyExecutioner) getActionArgs(ctx context.Context,
	dep *queuedDependency, depInstallation storage.Installation) error {
	finalParams, err := e.porter.finalizeParameters(ctx, depInstallation, dep.BundleReference.Definition, e.parentArgs.Action, dep.Parameters)
	if err != nil {
		return fmt.Errorf("error resolving parameters for dependency %s: %w", dep.Alias, err)
	}
	e.depArgs = cnabprovider.ActionArguments{
		BundleReference:       dep.BundleReference,
		Action:                e.parentArgs.Action,
		Installation:          depInstallation,
		Driver:                e.parentArgs.Driver,
		AllowDockerHostAccess: e.parentOpts.AllowDockerHostAccess,
		Params:                finalParams,
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
		//If we have depsv2 on, and no parentinstallation label, do not uninstall
		// this must be uninstalled directly
		if _, ok := e.depArgs.Installation.Labels["sh.porter.parentInstallation"]; !ok && dep.SharingMode {
			return nil
		}
		uninstallOpts = opts
	}

	//
	if e.parentAction.GetAction() == "upgrade" {
		if _, ok := e.depArgs.Installation.Labels["sh.porter.parentInstallation"]; !ok && dep.SharingMode {
			return nil
		}
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
	// will resolve to false and thus be a no-op
	if uninstallOpts.shouldDelete() {
		span.Infof(installationDeleteTmpl, e.depArgs.Installation)
		return e.Installations.RemoveInstallation(ctx, e.depArgs.Installation.Namespace, e.depArgs.Installation.Name)
	}
	return nil
}
