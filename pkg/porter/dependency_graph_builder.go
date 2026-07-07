package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
)

// GraphBuilder resolves the full, transitive dependency graph for a bundle.
type GraphBuilder struct {
	porter   *Porter
	maxDepth int
}

// NewGraphBuilder creates a new GraphBuilder.
func NewGraphBuilder(porter *Porter, maxDepth int) *GraphBuilder {
	return &GraphBuilder{porter: porter, maxDepth: maxDepth}
}

// BuildDependencyGraph resolves bun's full dependency graph: recursively
// pulling every referenced bundle, applying registry and version-strategy
// resolution at each level (so version ranges resolve the same way they do
// at install/upgrade time), deduplicating nodes that represent the same
// dependency instance, and recording both structural "requires" edges and
// data-flow "wiring" edges derived from output references between sibling
// v2 dependencies. Returns ErrDependencyCycle if the resulting graph isn't a
// DAG.
func (b *GraphBuilder) BuildDependencyGraph(ctx context.Context, bun cnab.ExtendedBundle, opts ExplainOpts) (*Graph, error) {
	g := newGraph()

	root := NodeKey{IsRoot: true}
	g.Root = root
	g.Nodes[root] = &Node{Key: root, Bundle: bun}

	if err := b.expandNode(ctx, g, root, bun, opts, 0); err != nil {
		return nil, err
	}

	if _, err := g.TopologicalOrder(); err != nil {
		return nil, err
	}

	return g, nil
}

// expandNode resolves the direct dependencies of bun (the bundle at graph
// node key), labels each with depth (0 for the root bundle's direct
// dependencies, 1 for their dependencies, and so on, matching how
// InspectableDependency.Depth has always been reported), and recurses.
//
// A node is added to g.Nodes before its own dependencies are expanded, so a
// cycle (a dependency that transitively requires or is wired back to an
// ancestor still being expanded) always finds that ancestor already present
// and simply stops recursing there rather than looping forever; the full
// cycle -- spanning both requires and wiring edges, with every participating
// node -- is then reported in one place by BuildDependencyGraph's final
// TopologicalOrder call, instead of duplicating a partial check here.
func (b *GraphBuilder) expandNode(
	ctx context.Context,
	g *Graph,
	key NodeKey,
	bun cnab.ExtendedBundle,
	opts ExplainOpts,
	depth int,
) error {
	if depth >= b.maxDepth {
		fmt.Fprintf(b.porter.Err, "warning: dependency graph exceeds max depth of %d, stopping traversal\n", b.maxDepth)
		return nil
	}

	if !bun.HasDependenciesV1() && !bun.HasDependenciesV2() {
		return nil
	}

	// Apply registry access and version-strategy resolution exactly like
	// dependencyExecutioner.identifyDependencies does, so version ranges
	// resolve the same way here as they do at install/upgrade time.
	strategy := opts.DependenciesVersionStrategy
	if strategy == "" {
		strategy = b.porter.GetDependenciesVersionStrategy()
	}
	regOpts := cnabtooci.RegistryOptions{InsecureRegistry: opts.InsecureRegistry}
	adapter := &registryListTagsAdapter{reg: b.porter.Registry, opts: regOpts}
	eb := bun.WithRegistry(adapter, regOpts).WithVersionStrategy(strategy)

	locks, err := eb.ResolveDependencies(ctx, bun)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Gather each dependency's wiring configuration (v2 only) up front, so
	// dedup keys and wiring edges can both be computed per-lock.
	var v2Requires map[string]v2.Dependency
	if bun.HasDependenciesV2() {
		v2Deps, err := bun.ReadDependenciesV2()
		if err != nil {
			return fmt.Errorf("failed to read v2 dependencies: %w", err)
		}
		v2Requires = v2Deps.Requires
	}

	// alias -> NodeKey, for resolving wiring refs (which are alias-scoped to
	// this level's Requires map) to actual node keys.
	aliasToKey := make(map[string]NodeKey, len(locks))

	for _, lock := range locks {
		var parameters, credentials map[string]string
		if dep, ok := v2Requires[lock.Alias]; ok {
			parameters = dep.Parameters
			credentials = dep.Credentials
		}

		childKey := computeNodeKey(lock.Reference, parameters, credentials, lock.SharingGroup)
		aliasToKey[lock.Alias] = childKey

		g.addEdge(Edge{
			From:         key,
			To:           childKey,
			Kind:         EdgeKindRequires,
			ToAlias:      lock.Alias,
			SharingMode:  lock.SharingMode,
			SharingGroup: lock.SharingGroup,
		})

		// A node already present means either a diamond (fully expanded
		// elsewhere) or a cycle back to an ancestor still being expanded --
		// either way, don't re-pull or re-recurse; see the doc comment above.
		if _, ok := g.Nodes[childKey]; ok {
			continue
		}

		node := &Node{Key: childKey, Depth: depth}
		g.Nodes[childKey] = node

		childBun, err := b.pullDependencyBundle(ctx, lock.Reference, opts)
		if err != nil {
			node.ResolutionFailed = true
			node.ResolutionError = err.Error()
			continue
		}
		node.Bundle = childBun

		if err := b.expandNode(ctx, g, childKey, childBun, opts, depth+1); err != nil {
			return err
		}
	}

	if v2Requires != nil {
		refs, dangling, invalid := extractWiringRefs(v2Requires)
		for _, ref := range refs {
			fromKey, ok := aliasToKey[ref.FromAlias]
			if !ok {
				continue
			}
			toKey, ok := aliasToKey[ref.ToAlias]
			if !ok {
				continue
			}
			detail := ref.Detail
			g.addEdge(Edge{From: fromKey, To: toKey, Kind: EdgeKindWiring, ToAlias: ref.ToAlias, Detail: &detail})
		}
		for _, d := range dangling {
			fromKey, ok := aliasToKey[d.FromAlias]
			if !ok {
				continue
			}
			node := g.Nodes[fromKey]
			if node == nil {
				continue
			}
			if d.SelfReference {
				node.Warnings = append(node.Warnings, fmt.Sprintf(
					"dependency %q references its own output %q, which is not available until after it runs",
					d.FromAlias, d.Detail.SourceOutput))
				continue
			}
			node.Warnings = append(node.Warnings, fmt.Sprintf(
				"dependency %q references output %q of unknown dependency %q",
				d.FromAlias, d.Detail.SourceOutput, d.ToAlias))
		}
		for _, inv := range invalid {
			fromKey, ok := aliasToKey[inv.FromAlias]
			if !ok {
				continue
			}
			if node := g.Nodes[fromKey]; node != nil {
				node.Warnings = append(node.Warnings, fmt.Sprintf(
					"dependency %q %s.%s references the root bundle's own output (%s), which cannot be used as a dependency's source",
					inv.FromAlias, inv.Field, inv.FieldName, inv.RawMatch))
			}
		}
	}

	return nil
}

// pullDependencyBundle pulls a dependency bundle from the registry.
func (b *GraphBuilder) pullDependencyBundle(ctx context.Context, ref string, opts ExplainOpts) (cnab.ExtendedBundle, error) {
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

// graphToInspectableDependencies renders g as the nested, depth-indented
// InspectableDependency shape porter inspect has always produced. A node
// reachable via multiple parents (a diamond dependency) is rendered once per
// occurrence, at each occurrence's own depth -- only the underlying
// resolution (one pull, one set of metadata) is deduplicated, not the
// display. Each node's subtree is computed once and reused (relabeled to
// the occurrence's depth) for every subsequent occurrence, so rendering a
// shared/diamond dependency graph stays proportional to the deduplicated
// graph rather than growing combinatorially with fan-out.
func graphToInspectableDependencies(g *Graph, parentKey NodeKey, depth int) []InspectableDependency {
	return renderInspectableDependencies(g, parentKey, depth, make(map[NodeKey][]InspectableDependency))
}

func renderInspectableDependencies(g *Graph, parentKey NodeKey, depth int, cache map[NodeKey][]InspectableDependency) []InspectableDependency {
	edges := g.EdgesFrom(parentKey)
	if len(edges) == 0 {
		return nil
	}

	parentNode := g.Nodes[parentKey]

	deps := make([]InspectableDependency, 0, len(edges))
	for _, edge := range edges {
		if edge.Kind != EdgeKindRequires {
			continue
		}

		child := g.Nodes[edge.To]
		if child == nil {
			continue
		}

		lock := cnab.DependencyLock{
			Alias:        edge.ToAlias,
			Reference:    child.Key.Reference,
			SharingMode:  edge.SharingMode,
			SharingGroup: edge.SharingGroup,
		}

		dep, err := buildInspectableDependency(parentNode.Bundle, lock, depth)
		if err != nil {
			// Metadata extraction failure: surface it the same way a pull
			// resolution failure is surfaced, rather than aborting the
			// whole graph rendering.
			deps = append(deps, InspectableDependency{
				Alias:            edge.ToAlias,
				Reference:        child.Key.Reference,
				Depth:            depth,
				SharingMode:      edge.SharingMode,
				SharingGroup:     edge.SharingGroup,
				ResolutionFailed: true,
				ResolutionError:  err.Error(),
			})
			continue
		}

		dep.ResolutionFailed = child.ResolutionFailed
		dep.ResolutionError = child.ResolutionError
		dep.Warnings = child.Warnings
		dep.WiringEdges = wiringEdgeSummaries(g, child.Key)

		if !child.ResolutionFailed {
			if cached, ok := cache[child.Key]; ok {
				dep.Dependencies = relabelInspectableDependencies(cached, depth+1)
			} else {
				childDeps := renderInspectableDependencies(g, child.Key, depth+1, cache)
				cache[child.Key] = childDeps
				dep.Dependencies = childDeps
			}
		}

		deps = append(deps, dep)
	}

	return deps
}

// relabelInspectableDependencies deep-copies a previously-rendered subtree
// and relabels every node's Depth relative to newDepth, so a cached subtree
// computed for one occurrence of a shared dependency can be reused at
// another occurrence reached at a different depth. Maps/slices on each
// dependency are shared with the cached original rather than copied, since
// nothing downstream mutates them after graphToInspectableDependencies
// returns.
func relabelInspectableDependencies(deps []InspectableDependency, newDepth int) []InspectableDependency {
	if len(deps) == 0 {
		return nil
	}

	relabeled := make([]InspectableDependency, len(deps))
	for i, dep := range deps {
		relabeled[i] = dep
		relabeled[i].Depth = newDepth
		relabeled[i].Dependencies = relabelInspectableDependencies(dep.Dependencies, newDepth+1)
	}
	return relabeled
}

// wiringEdgeSummaries returns a display-friendly summary of key's outgoing
// wiring edges to siblings under the same parent.
func wiringEdgeSummaries(g *Graph, key NodeKey) []WiringEdgeSummary {
	var summaries []WiringEdgeSummary
	for _, edge := range g.EdgesFrom(key) {
		if edge.Kind != EdgeKindWiring || edge.Detail == nil {
			continue
		}

		summaries = append(summaries, WiringEdgeSummary{
			Field:                 edge.Detail.Field,
			FieldName:             edge.Detail.FieldName,
			SourceDependencyAlias: edge.ToAlias,
			SourceOutput:          edge.Detail.SourceOutput,
		})
	}
	return summaries
}
