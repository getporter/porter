package porter

import (
	"fmt"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/cnab/extensions"
	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/pkg/errors"
)

type CNABAction func(cnabprovider.ActionArguments) error

type queuedDependency struct {
	extensions.DependencyLock
	CNABFile   string
	Parameters map[string]string
}

func (p *Porter) executeDependencies(parentOpts BundleLifecycleOpts, action CNABAction) error {
	// Load parent CNAB bundle definition
	var bun *bundle.Bundle
	if parentOpts.Tag != "" {
		var err error
		bunPath, _, _ := p.Cache.FindBundle(parentOpts.Tag)
		bun, err = p.CNAB.LoadBundle(bunPath, parentOpts.Insecure)
		if err != nil {
			return errors.Wrap(err, "could not load bundle from cache")
		}
	} else {
		bun, _ = p.CNAB.LoadBundle(parentOpts.CNABFile, parentOpts.Insecure)
	}

	solver := &extensions.DependencySolver{}
	locks, err := solver.ResolveDependencies(bun)
	if err != nil {
		return err
	}

	deps := make([]*queuedDependency, len(locks))
	for i, lock := range locks {
		deps[i] = &queuedDependency{
			DependencyLock: lock,
		}
	}

	// Prepare each dependency to flush out any problems before we run anything
	for _, dep := range deps {
		if p.Debug {
			fmt.Fprintf(p.Out, "Resolved dependency %s to %s\n", dep.Name, dep.Tag)
		}

		// Pull the dependency
		pullOpts := BundlePullOptions{
			Tag:              dep.Tag,
			InsecureRegistry: parentOpts.InsecureRegistry,
		}
		dep.CNABFile, err = p.PullBundle(pullOpts)
		if err != nil {
			return errors.Wrapf(err, "error pulling dependency %s", dep.Name)
		}

		// Load and validate it
		depBun, err := p.CNAB.LoadBundle(dep.CNABFile, parentOpts.Insecure)
		if err != nil {
			return errors.Wrapf(err, "could not load bundle %s", dep.Name)
		}

		// Make a lookup of which parameters are defined in the dependent bundle
		depParams := map[string]struct{}{}
		for paramName := range depBun.Parameters.Fields {
			depParams[paramName] = struct{}{}
		}

		// Handle any parameter overrides for the dependency defined in porter.yaml
		// dependencies:
		//   DEP:
		//     parameters:
		//       PARAM: VALUE
		for paramName, value := range p.Manifest.Dependencies[dep.Name].Parameters {
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
		for key, value := range parentOpts.combinedParameters {
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
				delete(parentOpts.combinedParameters, key)
				fmt.Println("added dependency param ", paramName)
			}
		}
	}

	// execute the requested action against all of the dependencies
	parentArgs := parentOpts.ToDuffleArgs()
	for _, dep := range deps {
		depArgs := cnabprovider.ActionArguments{
			Insecure:         parentArgs.Insecure,
			BundleIdentifier: dep.CNABFile,
			BundleIsFile:     false,
			Claim:            fmt.Sprintf("%s-%s", parentArgs.Claim, dep.Name),
			Driver:           parentOpts.Driver,
			Params:           dep.Parameters,

			// For now, assume it's okay to give the dependency the same credentials as the parent
			CredentialIdentifiers: parentArgs.CredentialIdentifiers,
		}
		fmt.Fprintf(p.Out, "Executing dependency %s...\n", dep.Name)
		err = action(depArgs)
		if err != nil {
			return errors.Wrapf(err, "error installing dependency %s", dep.Name)
		}
	}

	return nil
}
