package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	depsv1 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
)

// DependencyTreeBuilder builds the dependency tree for a bundle
type DependencyTreeBuilder struct {
	maxDepth int
	cache    map[string][]InspectableDependency
}

// NewDependencyTreeBuilder creates a new dependency tree builder
func NewDependencyTreeBuilder(maxDepth int) *DependencyTreeBuilder {
	return &DependencyTreeBuilder{
		maxDepth: maxDepth,
		cache:    make(map[string][]InspectableDependency),
	}
}

// BuildDependencyTree builds the dependency tree for a bundle
// Phase 2: Only builds direct dependencies (no recursion)
func (b *DependencyTreeBuilder) BuildDependencyTree(ctx context.Context, bun cnab.ExtendedBundle) ([]InspectableDependency, error) {
	// Resolve direct dependencies
	locks, err := bun.ResolveDependencies(bun)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	if len(locks) == 0 {
		return nil, nil
	}

	// Build inspectable dependencies for direct deps only
	deps := make([]InspectableDependency, 0, len(locks))
	for _, lock := range locks {
		dep, err := b.buildInspectableDependency(ctx, bun, lock, 0)
		if err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}

	return deps, nil
}

// buildInspectableDependency creates an InspectableDependency from a DependencyLock
func (b *DependencyTreeBuilder) buildInspectableDependency(
	ctx context.Context,
	bun cnab.ExtendedBundle,
	lock cnab.DependencyLock,
	depth int,
) (InspectableDependency, error) {
	dep := InspectableDependency{
		Alias:        lock.Alias,
		Reference:    lock.Reference,
		Depth:        depth,
		SharingMode:  lock.SharingMode,
		SharingGroup: lock.SharingGroup,
	}

	// Extract metadata based on v1 or v2
	if bun.HasDependenciesV2() {
		if err := b.extractV2Metadata(bun, &dep); err != nil {
			return dep, err
		}
	} else if bun.HasDependenciesV1() {
		if err := b.extractV1Metadata(bun, &dep); err != nil {
			return dep, err
		}
	}

	// Phase 2: No recursion yet
	// Phase 3 will pull the dependency bundle and recursively build its tree

	return dep, nil
}

// extractV1Metadata extracts metadata from v1 dependencies
func (b *DependencyTreeBuilder) extractV1Metadata(bun cnab.ExtendedBundle, dep *InspectableDependency) error {
	deps, err := bun.ReadDependenciesV1()
	if err != nil {
		return fmt.Errorf("failed to read v1 dependencies: %w", err)
	}

	v1Dep, ok := deps.Requires[dep.Alias]
	if !ok {
		return nil
	}

	// Extract version constraints
	dep.Version = formatVersionV1(v1Dep.Version)

	return nil
}

// extractV2Metadata extracts metadata from v2 dependencies
func (b *DependencyTreeBuilder) extractV2Metadata(bun cnab.ExtendedBundle, dep *InspectableDependency) error {
	deps, err := bun.ReadDependenciesV2()
	if err != nil {
		return fmt.Errorf("failed to read v2 dependencies: %w", err)
	}

	v2Dep, ok := deps.Requires[dep.Alias]
	if !ok {
		return nil
	}

	// Extract version constraints
	dep.Version = v2Dep.Version

	// Extract parameter wiring
	if len(v2Dep.Parameters) > 0 {
		dep.Parameters = make(map[string]string, len(v2Dep.Parameters))
		for k, v := range v2Dep.Parameters {
			dep.Parameters[k] = v
		}
	}

	// Extract credential wiring
	if len(v2Dep.Credentials) > 0 {
		dep.Credentials = make(map[string]string, len(v2Dep.Credentials))
		for k, v := range v2Dep.Credentials {
			dep.Credentials[k] = v
		}
	}

	// Extract output wiring
	if len(v2Dep.Outputs) > 0 {
		dep.Outputs = make(map[string]string, len(v2Dep.Outputs))
		for k, v := range v2Dep.Outputs {
			dep.Outputs[k] = v
		}
	}

	return nil
}

// formatVersionV1 formats v1 version constraints into a readable string
func formatVersionV1(ver *depsv1.DependencyVersion) string {
	if ver == nil || len(ver.Ranges) == 0 {
		return ""
	}

	// Join all ranges with " || "
	result := ""
	for i, r := range ver.Ranges {
		if i > 0 {
			result += " || "
		}
		result += r
	}

	if ver.AllowPrereleases {
		result += " (including prereleases)"
	}

	return result
}

// flattenDependencyTree flattens a nested dependency tree into a single list
// preserving the depth information for indentation
func flattenDependencyTree(deps []InspectableDependency) []InspectableDependency {
	result := make([]InspectableDependency, 0)
	for _, dep := range deps {
		result = append(result, dep)
		if len(dep.Dependencies) > 0 {
			nested := flattenDependencyTree(dep.Dependencies)
			result = append(result, nested...)
		}
	}
	return result
}
