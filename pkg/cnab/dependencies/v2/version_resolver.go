package v2

import (
	"context"
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/Masterminds/semver/v3"
)

var _ DependencyResolver = VersionResolver{}

// VersionResolver resolves the highest version of the default bundle reference of a Dependency.
type VersionResolver struct {
	puller BundlePuller
}

// Resolve attempts to find the highest available version of the default bundle for the specified Dependency.
//
// Returns the resolved bundle reference, whether a match was found, and an error if applicable.
func (v VersionResolver) ResolveDependency(ctx context.Context, dep Dependency) (Node, bool, error) {
	bundle := dep.DefaultBundle
	if bundle == nil || bundle.Version == nil {
		return nil, false, nil
	}

	tags, err := v.puller.ListTags(ctx, bundle.Reference)
	if err != nil {
		return nil, false, err
	}

	versions := make(semver.Collection, 0, len(tags))
	for _, tag := range tags {
		version, err := semver.NewVersion(tag)
		if err == nil {
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		return nil, false, nil
	}

	sort.Sort(sort.Reverse(versions))

	// TODO: return the first one that matches the bundle interface
	versionRef, err := bundle.Reference.WithTag(versions[0].Original())
	if err != nil {
		return nil, false, err
	}

	bunRef := cnab.BundleReference{Reference: versionRef}
	return BundleNode{
		Key:         dep.Key,
		ParentKey:   dep.ParentKey,
		Reference:   bunRef,
		Parameters:  dep.Parameters,
		Credentials: dep.Credentials,
	}, true, nil
}
