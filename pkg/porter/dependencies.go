package porter

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/runtime"
	"get.porter.sh/porter/pkg/storage"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type dependencyExecutioner struct {
	*context.Context
	porter *Porter

	Manifest *manifest.Manifest
	Resolver BundleResolver
	CNAB     cnabprovider.CNABProvider
	Claims   claims.Provider

	// These are populated by Prepare, call it or perish in inevitable errors
	parentAction BundleAction
	parentOpts   *BundleActionOptions
	parentArgs   cnabprovider.ActionArguments
	deps         []*queuedDependency
}

func newDependencyExecutioner(p *Porter, action BundleAction, actionArgs cnabprovider.ActionArguments) *dependencyExecutioner {
	resolver := BundleResolver{
		Cache:    p.Cache,
		Registry: p.Registry,
	}
	return &dependencyExecutioner{
		porter:       p,
		parentAction: action,
		parentOpts:   action.GetOptions(),
		parentArgs:   actionArgs,
		Context:      p.Context,
		Manifest:     p.Manifest,
		Resolver:     resolver,
		CNAB:         p.CNAB,
		Claims:       p.Claims,
	}
}

type queuedDependency struct {
	cnab.DependencyLock
	BundleReference cnab.BundleReference
	Parameters      map[string]interface{}

	// cache of the CNAB file contents
	cnabFileContents []byte
}

func (e *dependencyExecutioner) Prepare() error {
	err := e.identifyDependencies()
	if err != nil {
		return err
	}

	for _, dep := range e.deps {
		err := e.prepareDependency(dep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *dependencyExecutioner) Execute() error {
	if e.deps == nil {
		return errors.New("Prepare must be called before Execute")
	}

	// executeDependency the requested action against all of the dependencies
	for _, dep := range e.deps {
		err := e.executeDependency(dep)
		if err != nil {
			return err
		}
	}

	return nil
}

// PrepareRootActionArguments uses information about the dependencies of a bundle to prepare
// the execution of the root operation.
func (e *dependencyExecutioner) PrepareRootActionArguments(args *cnabprovider.ActionArguments) {
	if args.Files == nil {
		args.Files = make(map[string]string, 2*len(e.deps))
	}

	// Define files necessary for dependencies that need to be copied into the bundle
	// args.Files is a map of target path to file contents
	for _, dep := range e.deps {
		// Copy the dependency bundle.json
		target := runtime.GetDependencyDefinitionPath(dep.Alias)
		args.Files[target] = string(dep.cnabFileContents)
	}

	// Remove parameters for dependencies
	for key := range args.Params {
		if strings.Contains(key, "#") {
			delete(args.Params, key)
		}
	}
}

func (e *dependencyExecutioner) identifyDependencies() error {
	// Load parent CNAB bundle definition
	var bun cnab.ExtendedBundle
	if e.parentOpts.CNABFile != "" {
		bundle, err := e.CNAB.LoadBundle(e.parentOpts.CNABFile)
		if err != nil {
			return err
		}
		bun = bundle
	} else if e.parentOpts.Reference != "" {
		cachedBundle, err := e.Resolver.Resolve(e.parentOpts.BundlePullOptions)
		if err != nil {
			return errors.Wrapf(err, "could not resolve bundle")
		}

		bun = cachedBundle.Definition
	} else if e.parentOpts.Name != "" {
		c, err := e.Claims.GetLastRun(e.parentOpts.Namespace, e.parentOpts.Name)
		if err != nil {
			return err
		}

		bun = cnab.ExtendedBundle{c.Bundle}
	} else {
		// If we hit here, there is a bug somewhere
		return errors.New("identifyDependencies failed to load the bundle because no bundle was specified. Please report this bug to https://github.com/getporter/porter/issues/new/choose")
	}

	solver := &cnab.DependencySolver{}
	locks, err := solver.ResolveDependencies(bun)
	if err != nil {
		return err
	}

	e.deps = make([]*queuedDependency, len(locks))
	for i, lock := range locks {
		if e.Debug {
			fmt.Fprintf(e.Out, "Resolved dependency %s to %s\n", lock.Alias, lock.Reference)
		}
		e.deps[i] = &queuedDependency{
			DependencyLock: lock,
		}
	}

	return nil
}

func (e *dependencyExecutioner) prepareDependency(dep *queuedDependency) error {
	// Pull the dependency
	var err error
	pullOpts := BundlePullOptions{
		Reference:        dep.Reference,
		InsecureRegistry: e.parentOpts.InsecureRegistry,
		Force:            e.parentOpts.Force,
	}
	if err := pullOpts.Validate(); err != nil {
		return errors.Wrapf(err, "error preparing dependency %s", dep.Alias)
	}
	cachedDep, err := e.Resolver.Resolve(pullOpts)
	if err != nil {
		return errors.Wrapf(err, "error pulling dependency %s", dep.Alias)
	}
	dep.BundleReference = cachedDep.BundleReference

	err = cachedDep.Definition.Validate()
	if err != nil {
		return errors.Wrapf(err, "invalid bundle %s", dep.Alias)
	}

	// Cache the bundle.json for later
	dep.cnabFileContents, err = e.FileSystem.ReadFile(cachedDep.BundlePath)
	if err != nil {
		return errors.Wrapf(err, "error reading %s", cachedDep.BundlePath)
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
	for _, manifestDep := range e.Manifest.Dependencies.RequiredDependencies {
		if manifestDep.Name == dep.Alias {
			for paramName, value := range manifestDep.Parameters {
				// Make sure the parameter is defined in the bundle
				if _, ok := depParams[paramName]; !ok {
					return errors.Errorf("invalid dependencies.%s.parameters entry, %s is not a parameter defined in that bundle", dep.Alias, paramName)
				}

				if dep.Parameters == nil {
					dep.Parameters = make(map[string]interface{}, 1)
				}
				dep.Parameters[paramName] = value
			}
		}
	}

	// Handle any parameter overrides for the dependency defined on the command line
	// --param DEP#PARAM=VALUE
	for key, value := range e.parentArgs.Params {
		parts := strings.Split(key, "#")
		if len(parts) > 1 && parts[0] == dep.Alias {
			paramName := parts[1]

			// Make sure the parameter is defined in the bundle
			if _, ok := depParams[paramName]; !ok {
				return errors.Errorf("invalid --param %s, %s is not a parameter defined in the bundle %s", key, paramName, dep.Alias)
			}

			if dep.Parameters == nil {
				dep.Parameters = make(map[string]interface{}, 1)
			}
			dep.Parameters[paramName] = value
			delete(e.parentArgs.Params, key)
		}
	}

	return nil
}

func (e *dependencyExecutioner) executeDependency(dep *queuedDependency) error {
	// TODO(carolynvs): We should really switch up how the deperator works so that
	// even the root bundle uses the execution engine here. This would set up how
	// we want dependencies and mixins as bundles to work in the future.

	depName := cnab.BuildPrerequisiteInstallationName(e.parentOpts.Name, dep.Alias)
	depInstallation, err := e.Claims.GetInstallation(e.parentOpts.Namespace, depName)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound{}) {
			depInstallation = claims.NewInstallation(e.parentOpts.Namespace, depName)
			depInstallation.SetLabel("sh.porter.parentInstallation", e.parentArgs.Installation.String())
			if err = e.Claims.InsertInstallation(depInstallation); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	depArgs := cnabprovider.ActionArguments{
		BundleReference:       dep.BundleReference,
		Action:                e.parentArgs.Action,
		Installation:          depInstallation,
		Driver:                e.parentArgs.Driver,
		AllowDockerHostAccess: e.parentOpts.AllowAccessToDockerHost,
		Params:                dep.Parameters,
		// For now, assume it's okay to give the dependency the same credentials as the parent
		CredentialIdentifiers: e.parentArgs.CredentialIdentifiers,
	}

	// Determine if we're working with UninstallOptions, to inform deletion and
	// error handling, etc.
	var uninstallOpts UninstallOptions
	if opts, ok := e.parentAction.(UninstallOptions); ok {
		uninstallOpts = opts
	}

	var executeErrs error
	fmt.Fprintf(e.Out, "Executing dependency %s...\n", dep.Alias)
	err = e.CNAB.Execute(depArgs)
	if err != nil {
		executeErrs = multierror.Append(executeErrs, errors.Wrapf(err, "error executing dependency %s", dep.Alias))

		// Handle errors when/if the action is uninstall
		// If uninstallOpts is an empty struct, executeErrs will pass through
		executeErrs = uninstallOpts.handleUninstallErrs(e.Out, executeErrs)
		if executeErrs != nil {
			return executeErrs
		}
	}

	// If uninstallOpts is an empty struct (i.e., action not Uninstall), this
	// will resolve to false and thus be a no-op
	if uninstallOpts.shouldDelete() {
		fmt.Fprintf(e.Out, installationDeleteTmpl, depArgs.Installation)
		return e.Claims.RemoveInstallation(depArgs.Installation.Namespace, depArgs.Installation.Name)
	}
	return nil
}
