package dependencies

import (
	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/cnab/extensions"
	"github.com/pkg/errors"
)

func BuildDependencyQueue(bun *bundle.Bundle) ([]extensions.Dependency, error) {
	deps, err := extensions.LoadDependencies(bun)
	if deps == nil {
		return []extensions.Dependency{}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error executing dependencies for %s", bun.Name)
	}

	q := make([]extensions.Dependency, 0, len(deps.Requires))
	for _, dep := range deps.Requires {
		q = append(q, dep)
	}

	return q, nil
}
