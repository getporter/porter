package v2

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
)

var _ DependencyResolver = InstallationResolver{}

// InstallationResolver resolves an existing installation from a dependency
type InstallationResolver struct {
	store storage.InstallationProvider

	// Namespace of the root installation
	namespace string
}

// Resolve attempts to identify an existing installation that satisfies the
// specified Dependency.
//
// Returns the matching installation (if found), whether
// a matching installation was found, and an error if applicable.
func (r InstallationResolver) ResolveDependency(ctx context.Context, dep Dependency) (Node, bool, error) {
	if dep.InstallationSelector == nil {
		return nil, false, nil
	}

	// Build a query for matching installations
	filter := make(bson.M, 1)

	// Match installations with one of the specified namespaces
	namespacesQuery := make([]bson.M, 2)
	for _, ns := range dep.InstallationSelector.Namespaces {
		namespacesQuery = append(namespacesQuery, bson.M{"namespace": ns})
	}
	filter["$or"] = namespacesQuery

	// Match all specified labels
	for k, v := range dep.InstallationSelector.Labels {
		filter["labels."+k] = v
	}

	findOpts := storage.FindOptions{
		Sort:   []string{"-namespace", "name"},
		Filter: filter,
	}
	installations, err := r.store.FindInstallations(ctx, findOpts)
	if err != nil {
		return nil, false, err
	}

	// map[installation index]isMatchBool
	matches := make(map[int]bool)
	for i, inst := range installations {
		if dep.InstallationSelector.IsMatch(ctx, inst) {
			matches[i] = true
		}
	}

	switch len(matches) {
	case 0:
		return nil, false, nil
	case 1:
		var instIndex int
		for i := range matches {
			instIndex = i
		}
		inst := installations[instIndex]
		match := &InstallationNode{
			Key:       dep.Key,
			Namespace: inst.Namespace,
			Name:      inst.Name,
		}
		return match, true, nil
	default:
		var preferredMatch *storage.Installation
		// Prefer an installation that is the same as the default bundle if there are multiple interface matches
		if dep.DefaultBundle != nil {
			for i, isCandidate := range matches {
				if !isCandidate {
					continue
				}

				inst := installations[i]
				bundleRef, err := cnab.ParseOCIReference(inst.Status.BundleReference)
				if err != nil {
					matches[i] = false
					continue
				}

				if dep.DefaultBundle.Reference.Repository() == bundleRef.Repository() {
					preferredMatch = &inst
					break
				}

			}
		}

		// Prefer an installation in the same namespace if there is both a global and local installation
		if preferredMatch != nil && preferredMatch.Namespace == r.namespace {
			match := &InstallationNode{
				Key:       dep.Key,
				Namespace: preferredMatch.Namespace,
				Name:      preferredMatch.Name,
			}
			return match, true, nil
		}

		// Just pick the first installation sorted by -namespace, name (i.e. global last)
		for i, isCandidate := range matches {
			if !isCandidate {
				continue
			}

			inst := installations[i]
			match := &InstallationNode{
				Key:       dep.Key,
				Namespace: inst.Namespace,
				Name:      inst.Name,
			}
			return match, true, nil
		}

		return nil, false, nil
	}
}
