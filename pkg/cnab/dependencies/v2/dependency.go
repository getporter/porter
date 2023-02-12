package v2

import (
	"context"
	"fmt"

	depsv2ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/Masterminds/semver/v3"
	"github.com/cnabio/cnab-go/bundle"
	"go.opentelemetry.io/otel/attribute"
)

// Dependency is a fully hydrated representation of a bundle dependency with
// sufficient information to be resolved using a DependencyResolver to an action
// that can be used in an Execution Plan.
// TODO: can we come up with a better name, e.g. unresolve dependency, dependency selector, etc
type Dependency struct {
	Key                  string
	ParentKey            string
	DefaultBundle        *BundleReferenceSelector
	Interface            *BundleInterfaceSelector
	InstallationSelector *InstallationSelector
	Requires             []string
	Parameters           map[string]depsv2ext.DependencySource
	Credentials          map[string]depsv2ext.DependencySource
}

// BundleReferenceSelector evaluates the bundle criteria of a Dependency.
type BundleReferenceSelector struct {
	// Reference to a bundle, optionally including a default tag or digest.
	Reference cnab.OCIReference

	// Version specifies the range of allowed versions that Porter may select from
	// when resolving this bundle to a specific reference. When a version is not
	// specified or cannot be resolved, the tag/digest specified on the Reference is
	// used as a default.
	Version *semver.Constraints
}

// IsMatch determines if the specified installation satisfies a Dependency's bundle criteria.
func (s *BundleReferenceSelector) IsMatch(ctx context.Context, inst storage.Installation) bool {
	log := tracing.LoggerFromContext(ctx)
	log.Debug("Evaluating installation bundle definition")

	if inst.Status.BundleReference == "" {
		log.Debug("Installation does not match because it does not have an associated bundle")
		return false
	}

	ref, err := cnab.ParseOCIReference(inst.Status.BundleReference)
	if err != nil {
		log.Warn("Could not evaluate installation because the BundleReference is invalid",
			attribute.String("reference", inst.Status.BundleReference))
		return false
	}

	// If no selector is defined, consider it a match
	if s == nil {
		return true
	}

	// If a version range is specified, ignore the version on the selector and apply the range
	// otherwise match the tag or digest
	if s.Version != nil {
		if inst.Status.BundleVersion == "" {
			log.Debug("Installation does not match because it does not have an associated bundle version")
			return false
		}

		// First check that the repository is the same
		gotRepo := ref.Repository()
		wantRepo := s.Reference.Repository()
		if gotRepo != wantRepo {
			log.Warn("Installation does not match because the bundle repository is incorrect",
				attribute.String("installation-bundle-repository", gotRepo),
				attribute.String("dependency-bundle-repository", wantRepo),
			)
			return false
		}

		gotVersion, err := semver.NewVersion(inst.Status.BundleVersion)
		if err != nil {
			log.Warn("Installation does not match because the bundle version is invalid",
				attribute.String("installation-bundle-version", inst.Status.BundleVersion),
			)
			return false
		}

		if s.Version.Check(gotVersion) {
			log.Debug("Installation matches because the bundle version is in range",
				attribute.String("installation-bundle-version", inst.Status.BundleVersion),
				attribute.String("dependency-bundle-version", s.Version.String()),
			)
			return true
		} else {
			log.Debug("Installation does not match because the bundle version is incorrect",
				attribute.String("installation-bundle-version", inst.Status.BundleVersion),
				attribute.String("dependency-bundle-version", s.Version.String()),
			)
			return false
		}
	} else {
		gotRef := ref.String()
		wantRef := s.Reference.String()
		if gotRef == wantRef {
			log.Warn("Installation matches because the bundle reference is correct",
				attribute.String("installation-bundle-reference", gotRef),
				attribute.String("dependency-bundle-reference", wantRef),
			)
			return true
		} else {
			log.Warn("Installation does not match because the bundle reference is incorrect",
				attribute.String("installation-bundle-reference", gotRef),
				attribute.String("dependency-bundle-reference", wantRef),
			)
			return false
		}
	}
}

// InstallationSelector evaluates the installation criteria of a Dependency.
type InstallationSelector struct {
	// Bundle is the criteria used for evaluating if a bundle satisfies a Dependency.
	Bundle *BundleReferenceSelector

	// Interface is the criteria used for evaluating if an installation or bundle
	// satisfies a Dependency.
	Interface *BundleInterfaceSelector

	// Labels is the set of labels used to find an existing installation that may be
	// used to satisfy a Dependency.
	Labels map[string]string

	// Namespaces is the set of namespaces used when searching for an existing
	// installation that may be used to satisfy a Dependency.
	Namespaces []string
}

// IsMatch determines if the specified installation satisfies a Dependency's installation criteria.
func (s InstallationSelector) IsMatch(ctx context.Context, inst storage.Installation) bool {
	// Skip checking labels and namespaces, those were used to query the set of
	// installations that we are checking

	bundleMatches := s.Bundle.IsMatch(ctx, inst)
	if !bundleMatches {
		return false
	}

	interfaceMatches := s.Interface.IsMatch(ctx, inst)
	return interfaceMatches
}

// BundleInterfaceSelector defines how a bundle is going to be used.
// It is not the same as the bundle definition.
// It works like go interfaces where its defined by its consumer.
type BundleInterfaceSelector struct {
	Parameters  []bundle.Parameter
	Credentials []bundle.Credential
	Outputs     []bundle.Output
}

// IsMatch determines if the specified installation satisfies a Dependency's bundle interface criteria.
func (s BundleInterfaceSelector) IsMatch(ctx context.Context, inst storage.Installation) bool {
	// TODO: implement
	return true
}

func MakeDependencyKey(parent string, dep string) string {
	return fmt.Sprintf("%s/%s", parent, dep)
}
