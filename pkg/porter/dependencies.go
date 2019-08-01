package porter

import (
	"fmt"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/cnab/extensions"
	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

type dependencyExecutioner struct {
	*context.Context
	Resolver BundleResolver
	CNAB     CNABProvider
	Manifest *config.Manifest

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
		Context:  p.Context,
		Resolver: resolver,
		CNAB:     p.CNAB,
		Manifest: p.Manifest,
	}
}

type cnabAction func(cnabprovider.ActionArguments) error

type queuedDependency struct {
	extensions.DependencyLock
	CNABFile   string
	Parameters map[string]string
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

func (e *dependencyExecutioner) Execute() error {
	if e.action == nil {
		return errors.New("Prepare must be called before Execute")
	}

	// executeDependency the requested action against all of the dependencies
	parentArgs := e.parentOpts.ToDuffleArgs()
	for _, dep := range e.deps {
		err := e.executeDependency(dep, parentArgs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *dependencyExecutioner) identifyDependencies() error {
	// Load parent CNAB bundle definition
	var bun *bundle.Bundle
	if e.parentOpts.Tag != "" {
		bunPath, err := e.Resolver.Resolve(e.parentOpts.BundlePullOptions)
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
			fmt.Fprintf(e.Out, "Resolved dependency %s to %s\n", lock.Name, lock.Tag)
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
	}
	dep.CNABFile, err = e.Resolver.Resolve(pullOpts)
	if err != nil {
		return errors.Wrapf(err, "error pulling dependency %s", dep.Name)
	}

	// Load and validate it
	depBun, err := e.CNAB.LoadBundle(dep.CNABFile, e.parentOpts.Insecure)
	if err != nil {
		return errors.Wrapf(err, "could not load bundle %s", dep.Name)
	}

	err = depBun.Validate()
	if err != nil {
		return errors.Wrapf(err, "invalid bundle %s", dep.Name)
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
	for paramName, value := range e.Manifest.Dependencies[dep.Name].Parameters {
		// Make sure the parameter is defined in the bundle
		if _, ok := depParams[paramName]; !ok {
			return errors.Errorf("invalid dependencies.%s.parameters entry, %s is not a parameter defined in that bundle", dep.Name, paramName)
		}

		if dep.Parameters == nil {
			dep.Parameters = make(map[string]string, 1)
		}
		dep.Parameters[paramName] = value
	}

	// Handle any parameter overrides for the dependency defined on the command line
	// --param DEP#PARAM=VALUE
	for key, value := range e.parentOpts.combinedParameters {
		parts := strings.Split(key, "#")
		if len(parts) > 1 && parts[0] == dep.Name {
			fmt.Println(key)
			paramName := parts[1]

			// Make sure the parameter is defined in the bundle
			if _, ok := depParams[paramName]; !ok {
				return errors.Errorf("invalid --param %s, %s is not a parameter defined in the bundle %s", key, paramName, dep.Name)
			}

			if dep.Parameters == nil {
				dep.Parameters = make(map[string]string, 1)
			}
			dep.Parameters[paramName] = value
			delete(e.parentOpts.combinedParameters, key)
			fmt.Println("added dependency param ", paramName)
		}
	}
	return nil
}

func (e *dependencyExecutioner) executeDependency(dep *queuedDependency, parentArgs cnabprovider.ActionArguments) error {
	depArgs := cnabprovider.ActionArguments{
		Insecure:         parentArgs.Insecure,
		BundleIdentifier: dep.CNABFile,
		BundleIsFile:     false,
		Claim:            fmt.Sprintf("%s-%s", parentArgs.Claim, dep.Name),
		Driver:           parentArgs.Driver,
		Params:           dep.Parameters,

		// For now, assume it's okay to give the dependency the same credentials as the parent
		CredentialIdentifiers: parentArgs.CredentialIdentifiers,
	}
	fmt.Fprintf(e.Out, "Executing dependency %s...\n", dep.Name)
	err := e.action(depArgs)
	if err != nil {
		return errors.Wrapf(err, "error installing dependency %s", dep.Name)
	}
	return nil
}
