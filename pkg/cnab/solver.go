package cnab

import (
	"fmt"
	"sort"

	depsv1 "get.porter.sh/porter/pkg/cnab/dependencies/v1"
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/crane"
)

type DependencyLock struct {
	Alias     string
	Reference string
}

// TODO: move this logic onto the new ExtendedBundle struct
type DependencySolver struct {
}

func (s *DependencySolver) ResolveDependencies(bun ExtendedBundle) ([]DependencyLock, error) {
	if !bun.HasDependenciesV1() {
		return nil, nil
	}

	rawDeps, err := bun.ReadDependenciesV1()
	// We need make sure the DependenciesV1 are ordered by the desired sequence
	orderedDeps := rawDeps.ListBySequence()

	if err != nil {
		return nil, fmt.Errorf("error executing dependencies for %s: %w", bun.Name, err)
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
func (s *DependencySolver) ResolveVersion(name string, dep depsv1.Dependency) (OCIReference, error) {
	ref, err := ParseOCIReference(dep.Bundle)
	if err != nil {
		return OCIReference{}, fmt.Errorf("error parsing dependency (%s) bundle %q as OCI reference: %w", name, dep.Bundle, err)
	}

	// Here is where we could split out this logic into multiple strategy funcs / structs if necessary
	if dep.Version == nil || len(dep.Version.Ranges) == 0 {
		// Check if they specified an explicit tag in referenced bundle already
		if ref.HasTag() {
			return ref, nil
		}

		tag, err := s.determineDefaultTag(dep)
		if err != nil {
			return OCIReference{}, err
		}

		return ref.WithTag(tag)
	}

	return OCIReference{}, fmt.Errorf("not implemented: dependency version range specified for %s: %w", name, err)
}

func (s *DependencySolver) determineDefaultTag(dep depsv1.Dependency) (string, error) {
	tags, err := crane.ListTags(dep.Bundle)
	if err != nil {
		return "", fmt.Errorf("error listing tags for %s: %w", dep.Bundle, err)
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
			return "", fmt.Errorf("no tag was specified for %s and none of the tags defined in the registry meet the criteria: semver formatted or 'latest'", dep.Bundle)
		}
	}

	sort.Sort(sort.Reverse(versions))

	return versions[0].Original(), nil
}
