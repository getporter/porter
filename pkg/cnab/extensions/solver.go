package extensions

import (
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/Masterminds/semver/v3"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/pkg/errors"
)

type DependencyLock struct {
	Alias     string
	Reference string
}

type DependencySolver struct {
}

func (s *DependencySolver) ResolveDependencies(bun bundle.Bundle) ([]DependencyLock, error) {
	if !HasDependencies(bun) {
		return nil, nil
	}

	rawDeps, err := ReadDependencies(bun)
	// We need make sure the Dependencies are ordered by the desired sequence
	orderedDeps := rawDeps.ListBySequence()

	if err != nil {
		return nil, errors.Wrapf(err, "error executing dependencies for %s", bun.Name)
	}

	q := make([]DependencyLock, 0, len(orderedDeps))
	for _, dep := range orderedDeps {
		ref, err := s.ResolveVersion(dep.Name, dep)
		if err != nil {
			return nil, err
		}

		lock := DependencyLock{
			Alias:     dep.Name,
			Reference: ref.String(),
		}
		q = append(q, lock)
	}

	return q, nil
}

// ResolveVersion returns the bundle name, its version and any error.
func (s *DependencySolver) ResolveVersion(name string, dep Dependency) (cnab.OCIReference, error) {
	ref, err := cnab.ParseOCIReference(dep.Bundle)
	if err != nil {
		return cnab.OCIReference{}, errors.Wrapf(err, "error parsing dependency (%s) bundle %q as OCI reference", name, dep.Bundle)
	}

	// Here is where we could split out this logic into multiple strategy funcs / structs if necessary
	if dep.Version == nil || len(dep.Version.Ranges) == 0 {
		// Check if they specified an explicit tag in referenced bundle already
		if ref.HasTag() {
			return ref, nil
		}

		tag, err := s.determineDefaultTag(dep)
		if err != nil {
			return cnab.OCIReference{}, err
		}

		return ref.WithTag(tag)
	}

	return cnab.OCIReference{}, errors.Errorf("not implemented: dependency version range specified for %s", name)
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
