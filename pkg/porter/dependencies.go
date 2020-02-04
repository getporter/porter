package porter

import (
	"fmt"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab/extensions"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/runtime"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

type dependencyExecutioner struct {
	*context.Context
	// See https://github.com/deislabs/porter/issues/799
	// Manifest        *manifest.Manifest
	Resolver BundleResolver
	CNAB     CNABProvider
	Claims   claims.ClaimProvider

	// These are populated by Prepare, call it or perish in inevitable errors
	parentOpts BundleLifecycleOpts
	action     cnabAction
	deps       []*queuedDependency
}

func newDependencyExecutioner(p *Porter) *dependencyExecutioner {
	resolver := BundleResolver{
		Cache:    p.Cache,
		Registry: p.Registry,
	}
	return &dependencyExecutioner{
		Context: p.Context,
		// See https://github.com/deislabs/porter/issues/799
		// Manifest:        p.Manifest,
		Resolver: resolver,
		CNAB:     p.CNAB,
		Claims:   p.Claims,
	}
}

type cnabAction func(cnabprovider.ActionArguments) error

type queuedDependency struct {
	extensions.DependencyLock
	CNABFile   string
	Parameters map[string]string

	outputs map[string]interface{}

	// cache of the CNAB file contents
	cnabFileContents []byte

	RelocationMapping string
}

func (e *dependencyExecutioner) Prepare(parentOpts BundleLifecycleOpts, action cnabAction) error {
	e.parentOpts = parentOpts
	e.action = action

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

func (e *dependencyExecutioner) Execute(action manifest.Action) error {
	if e.action == nil {
		return errors.New("Prepare must be called before Execute")
	}

	// executeDependency the requested action against all of the dependencies
	parentArgs := e.parentOpts.ToActionArgs(e)
	for _, dep := range e.deps {
		err := e.executeDependency(dep, parentArgs, action)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *dependencyExecutioner) ApplyDependencyMappings(args *cnabprovider.ActionArguments) {
	if args.Files == nil {
		args.Files = make(map[string]string, 2*len(e.deps))
	}

	// Define files necessary for dependencies that need to be copied into the bundle
	// args.Files is a map of target path to file contents
	for _, dep := range e.deps {
		// Copy the dependency bundle.json
		target := runtime.GetDependencyDefinitionPath(dep.Alias)
		args.Files[target] = string(dep.cnabFileContents)

		// Copy the dependency output files defined from the bundle.json (loaded in prepareDependency)
		for output, value := range dep.outputs {
			target := filepath.Join(runtime.GetDependencyOutputsDir(dep.Alias), filepath.Base(output))
			args.Files[target] = fmt.Sprintf("%v", value)
		}
	}
}

func (e *dependencyExecutioner) identifyDependencies() error {
	// Load parent CNAB bundle definition
	var bun *bundle.Bundle
	if e.parentOpts.Tag != "" {
		bunPath, _, err := e.Resolver.Resolve(e.parentOpts.BundlePullOptions)
		if err != nil {
			return errors.Wrapf(err, "could not resolve bundle")
		}

		bun, err = e.CNAB.LoadBundle(bunPath, e.parentOpts.Insecure)
		if err != nil {
			return errors.Wrap(err, "could not load bundle from cache")
		}
	} else {
		bun, _ = e.CNAB.LoadBundle(e.parentOpts.CNABFile, e.parentOpts.Insecure)
	}

	solver := &extensions.DependencySolver{}
	locks, err := solver.ResolveDependencies(bun)
	if err != nil {
		return err
	}

	e.deps = make([]*queuedDependency, len(locks))
	for i, lock := range locks {
		if e.Debug {
			fmt.Fprintf(e.Out, "Resolved dependency %s to %s\n", lock.Alias, lock.Tag)
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
		Tag:              dep.Tag,
		InsecureRegistry: e.parentOpts.InsecureRegistry,
		Force:            e.parentOpts.Force,
	}
	dep.CNABFile, dep.RelocationMapping, err = e.Resolver.Resolve(pullOpts)
	if err != nil {
		return errors.Wrapf(err, "error pulling dependency %s", dep.Alias)
	}

	// Load and validate it
	depBun, err := e.CNAB.LoadBundle(dep.CNABFile, e.parentOpts.Insecure)
	if err != nil {
		return errors.Wrapf(err, "could not load bundle %s", dep.Alias)
	}

	err = depBun.Validate()
	if err != nil {
		return errors.Wrapf(err, "invalid bundle %s", dep.Alias)
	}

	// Cache the bundle.json for later
	dep.cnabFileContents, err = e.FileSystem.ReadFile(dep.CNABFile)
	if err != nil {
		return errors.Wrapf(err, "error reading %s", dep.CNABFile)
	}

	// Make a lookup of which parameters are defined in the dependent bundle
	depParams := map[string]struct{}{}
	for paramName := range depBun.Parameters {
		depParams[paramName] = struct{}{}
	}

	// Handle any parameter overrides for the dependency defined in porter.yaml
	// dependencies:
	//   DEP:
	//     parameters:
	//       PARAM: VALUE
	/* Disabling this until we have access to the manifest when working with a bundle from a tag
	   See https://github.com/deislabs/porter/issues/799
	for paramName, value := range e.Manifest.Dependencies[dep.Alias].Parameters {
		// Make sure the parameter is defined in the bundle
		if _, ok := depParams[paramName]; !ok {
			return errors.Errorf("invalid dependencies.%s.parameters entry, %s is not a parameter defined in that bundle", dep.Alias, paramName)
		}

		if dep.Parameters == nil {
			dep.Parameters = make(map[string]string, 1)
		}
		dep.Parameters[paramName] = value
	}
	*/

	// Handle any parameter overrides for the dependency defined on the command line
	// --param DEP#PARAM=VALUE
	for key, value := range e.parentOpts.combinedParameters {
		parts := strings.Split(key, "#")
		if len(parts) > 1 && parts[0] == dep.Alias {
			paramName := parts[1]

			// Make sure the parameter is defined in the bundle
			if _, ok := depParams[paramName]; !ok {
				return errors.Errorf("invalid --param %s, %s is not a parameter defined in the bundle %s", key, paramName, dep.Alias)
			}

			if dep.Parameters == nil {
				dep.Parameters = make(map[string]string, 1)
			}
			dep.Parameters[paramName] = value
			delete(e.parentOpts.combinedParameters, key)
		}
	}

	return nil
}

func (e *dependencyExecutioner) executeDependency(dep *queuedDependency, parentArgs cnabprovider.ActionArguments, action manifest.Action) error {
	depArgs := cnabprovider.ActionArguments{
		Insecure:          parentArgs.Insecure,
		BundlePath:        dep.CNABFile,
		Claim:             fmt.Sprintf("%s-%s", parentArgs.Claim, dep.Alias),
		Driver:            parentArgs.Driver,
		Params:            dep.Parameters,
		RelocationMapping: dep.RelocationMapping,

		// For now, assume it's okay to give the dependency the same credentials as the parent
		CredentialIdentifiers: parentArgs.CredentialIdentifiers,
	}
	fmt.Fprintf(e.Out, "Executing dependency %s...\n", dep.Alias)
	err := e.action(depArgs)
	if err != nil {
		return errors.Wrapf(err, "error executing dependency %s", dep.Alias)
	}

	// If action is uninstall, no claim will exist
	if action != manifest.ActionUninstall {
		// Collect expected outputs via claim
		c, err := e.Claims.Read(depArgs.Claim)
		if err != nil {
			return err
		}

		dep.outputs = c.Outputs
	}

	return nil
}
