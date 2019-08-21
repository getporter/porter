package extensions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/deislabs/cnab-go/bundle"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/pkg/errors"
)

type DependencyLock struct {
	Alias string
	Tag   string
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
	for alias, dep := range deps.Requires {
		bundle := strings.Split(dep.Bundle, ":")[0]
		version, err := s.ResolveVersion(alias, dep)
		if err != nil {
			return nil, err
		}
		lock := DependencyLock{
			Alias: alias,
			Tag:   fmt.Sprintf("%s:%s", bundle, version),
		}
		q = append(q, lock)
	}

	return q, nil
}

func (s *DependencySolver) ResolveVersion(alias string, dep Dependency) (string, error) {
	// Here is where we could split out this logic into multiple strategy funcs / structs if necessary
	if dep.Version == nil || len(dep.Version.Ranges) == 0 {
		parts := strings.Split(dep.Bundle, ":")
		if len(parts) > 1 {
			return strings.Join(parts[1:], ""), nil
		} else {
			return s.determineDefaultTag(dep)
		}
	}

	return "", errors.Errorf("not implemented: dependency version range specified for %s", alias)
}

func (s *DependencySolver) determineDefaultTag(dep Dependency) (string, error) {
	tags, err := crane.ListTags(dep.Bundle)
	if err != nil {
		return "", errors.Wrapf(err, "error listing tags for %s", dep.Bundle)
	}

	allowPrereleases := false
	if dep.Version != nil && dep.Version.AllowPrereleases {
		allowPrereleases = true
	}

	var hasLatest bool
	versions := make(semver.Collection, 0, len(tags))
	for _, tag := range tags {
		if tag == "latest" {
			hasLatest = true
			continue
		}

		version, err := semver.NewVersion(tag)
		if err == nil {
			if !allowPrereleases && version.Prerelease() != "" {
				continue
			}
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		if hasLatest {
			return "latest", nil
		} else {
			return "", errors.Errorf("no tag was specified for %s and none of the tags defined in the registry meet the criteria: semver formatted or 'latest'", dep.Bundle)
		}
	}

	sort.Sort(sort.Reverse(versions))

	return versions[0].Original(), nil
}
