package porter

import (
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/cnab/extensions"
	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/pkg/errors"
)

type CNABAction func(cnabprovider.ActionArguments) error

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
	deps, err := solver.ResolveDependencies(bun)
	if err != nil {
		return err
	}

	// Remember where the deps are pulled to
	cachedDeps := make(map[string]string, len(deps))

	// pre-pull each dependency to flush out any problems accessing them before we run anything
	for _, dep := range deps {
		if p.Debug {
			fmt.Fprintf(p.Out, "Resolved dependency %s to %s\n", dep.Name, dep.Tag)
		}
		pullOpts := BundlePullOptions{
			Tag:              dep.Tag,
			InsecureRegistry: parentOpts.InsecureRegistry,
		}
		cachePath, err := p.PullBundle(pullOpts)
		cachedDeps[dep.Name] = cachePath
		if err != nil {
			return errors.Wrapf(err, "error pulling dependency %s", dep.Name)
		}
	}

	// execute the requested action against all of the dependencies
	parentArgs := parentOpts.ToDuffleArgs()
	for _, dep := range deps {
		depArgs := cnabprovider.ActionArguments{
			Insecure:         parentArgs.Insecure,
			BundleIdentifier: cachedDeps[dep.Name],
			BundleIsFile:     false,
			Claim:            fmt.Sprintf("%s-%s", parentArgs.Claim, dep.Name),
			Driver:           parentOpts.Driver,
			// TODO: Provide credentials and parameters
			Params:                map[string]string{"mysql-user": "wordpress"},
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
