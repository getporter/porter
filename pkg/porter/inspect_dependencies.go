package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	depsv1 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
)

// DependencyTreeBuilder builds the dependency tree for a bundle
type DependencyTreeBuilder struct {
	porter   *Porter
	maxDepth int
	cache    map[string][]InspectableDependency
}

// NewDependencyTreeBuilder creates a new dependency tree builder
func NewDependencyTreeBuilder(porter *Porter, maxDepth int) *DependencyTreeBuilder {
	return &DependencyTreeBuilder{
		porter:   porter,
		maxDepth: maxDepth,
		cache:    make(map[string][]InspectableDependency),
	}
}

// BuildDependencyTree builds the dependency tree for a bundle
func (b *DependencyTreeBuilder) BuildDependencyTree(ctx context.Context, bun cnab.ExtendedBundle, opts ExplainOpts) ([]InspectableDependency, error) {
	visited := make(map[string]bool)
	return b.buildTreeRecursive(ctx, bun, opts, 0, visited)
}

// buildTreeRecursive recursively builds the dependency tree
func (b *DependencyTreeBuilder) buildTreeRecursive(
	ctx context.Context,
	bun cnab.ExtendedBundle,
	opts ExplainOpts,
	depth int,
	visited map[string]bool,
) ([]InspectableDependency, error) {
	// Check max depth
	if depth >= b.maxDepth {
		fmt.Fprintf(b.porter.Err, "warning: dependency tree exceeds max depth of %d, stopping traversal\n", b.maxDepth)
		return nil, nil
	}

	// Resolve direct dependencies
	locks, err := bun.ResolveDependencies(ctx, bun)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	if len(locks) == 0 {
		return nil, nil
	}

	// Build inspectable dependencies
	deps := make([]InspectableDependency, 0, len(locks))
	for _, lock := range locks {
		// Check for cycles
		if visited[lock.Reference] {
			fmt.Fprintf(b.porter.Err, "warning: circular dependency detected: %s\n", lock.Reference)
			continue
		}

		// Check cache
		if cached, ok := b.cache[lock.Reference]; ok {
			// Clone cached result to avoid mutation
			cloned := cloneDependencyTree(cached)
			// Update depth for cloned dependencies
			updateDepth(cloned, depth)
			deps = append(deps, cloned...)
			continue
		}

		// Build inspectable dependency with metadata
		dep, err := b.buildInspectableDependency(ctx, bun, lock, depth)
		if err != nil {
			return nil, err
		}

		// Pull dependency bundle and recurse
		childBun, err := b.pullDependencyBundle(ctx, lock.Reference, opts)
		if err != nil {
			// Mark dependency as failed and store error for verbose output
			dep.ResolutionFailed = true
			dep.ResolutionError = err.Error()
			// Continue with empty nested dependencies
			dep.Dependencies = []InspectableDependency{}
		} else {
			// Mark as visited
			visited[lock.Reference] = true

			// Recurse on children
			childDeps, err := b.buildTreeRecursive(ctx, childBun, opts, depth+1, visited)
			if err != nil {
				return nil, err
			}
			dep.Dependencies = childDeps

			// Unmark visited (allow in other branches)
			delete(visited, lock.Reference)

			// Cache result
			b.cache[lock.Reference] = childDeps
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

	return dep, nil
}

// pullDependencyBundle pulls a dependency bundle from the registry
func (b *DependencyTreeBuilder) pullDependencyBundle(
	ctx context.Context,
	ref string,
	opts ExplainOpts,
) (cnab.ExtendedBundle, error) {
	pullOpts := BundlePullOptions{
		Reference:        ref,
		InsecureRegistry: opts.InsecureRegistry,
		Force:            false,
	}

	resolver := BundleResolver{
		Cache:    b.porter.Cache,
		Registry: b.porter.Registry,
	}

	cachedBundle, err := resolver.Resolve(ctx, pullOpts)
	if err != nil {
		return cnab.ExtendedBundle{}, fmt.Errorf("failed to pull bundle %s: %w", ref, err)
	}

	return cachedBundle.Definition, nil
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

// cloneDependencyTree creates a deep copy of a dependency tree
func cloneDependencyTree(deps []InspectableDependency) []InspectableDependency {
	if len(deps) == 0 {
		return nil
	}

	cloned := make([]InspectableDependency, len(deps))
	for i, dep := range deps {
		cloned[i] = InspectableDependency{
			Alias:            dep.Alias,
			Reference:        dep.Reference,
			Version:          dep.Version,
			Depth:            dep.Depth,
			SharingMode:      dep.SharingMode,
			SharingGroup:     dep.SharingGroup,
			ResolutionFailed: dep.ResolutionFailed,
			ResolutionError:  dep.ResolutionError,
		}

		// Clone maps
		if len(dep.Parameters) > 0 {
			cloned[i].Parameters = make(map[string]string, len(dep.Parameters))
			for k, v := range dep.Parameters {
				cloned[i].Parameters[k] = v
			}
		}
		if len(dep.Credentials) > 0 {
			cloned[i].Credentials = make(map[string]string, len(dep.Credentials))
			for k, v := range dep.Credentials {
				cloned[i].Credentials[k] = v
			}
		}
		if len(dep.Outputs) > 0 {
			cloned[i].Outputs = make(map[string]string, len(dep.Outputs))
			for k, v := range dep.Outputs {
				cloned[i].Outputs[k] = v
			}
		}

		// Recursively clone nested dependencies
		if len(dep.Dependencies) > 0 {
			cloned[i].Dependencies = cloneDependencyTree(dep.Dependencies)
		}
	}

	return cloned
}

// updateDepth updates the depth of all dependencies in a tree
func updateDepth(deps []InspectableDependency, newDepth int) {
	for i := range deps {
		deps[i].Depth = newDepth
		if len(deps[i].Dependencies) > 0 {
			updateDepth(deps[i].Dependencies, newDepth+1)
		}
	}
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
