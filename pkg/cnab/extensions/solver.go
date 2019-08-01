package extensions

import (
	"fmt"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/pkg/errors"
)

type DependencyLock struct {
	Name string
	Tag  string
}

type DependencySolver struct {
}

func (s *DependencySolver) ResolveDependencies(bun *bundle.Bundle) ([]DependencyLock, error) {
	deps, err := ReadDependencies(bun)
	if deps == nil {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error executing dependencies for %s", bun.Name)
	}

	q := make([]DependencyLock, 0, len(deps.Requires))
	for name, dep := range deps.Requires {
		bundle := strings.Split(dep.Bundle, ":")[0]
		version, err := s.ResolveVersion(name, dep)
		if err != nil {
			return nil, err
		}
		lock := DependencyLock{
			Name: name,
			Tag:  fmt.Sprintf("%s:%s", bundle, version),
		}
		q = append(q, lock)
	}

	return q, nil
}

func (s *DependencySolver) ResolveVersion(name string, dep Dependency) (string, error) {
	// Here is where we could split out this logic into multiple strategy funcs / structs if necessary
	if dep.Version == nil {
		parts := strings.Split(dep.Bundle, ":")
		if len(parts) > 1 {
			return strings.Join(parts[1:], ""), nil
		} else {
			return "", errors.Errorf("not implemented: unspecified dependency version for %s", name)
		}
	}

	return "", errors.Errorf("not implemented: dependency version range specified for %s", name)
}
