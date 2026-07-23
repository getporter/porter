package porter

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"get.porter.sh/porter/pkg/experimental"
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

	ancestors := make(map[NodeKey]NodeKey)
	if err := b.expandNode(ctx, g, root, bun, opts, 0, ancestors); err != nil {
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
// ancestors maps the contentKey of every node currently being expanded (on
// this call's recursion stack) to the NodeKey it was assigned, regardless of
// whether that node is shareable. A dependency whose content matches an
// ancestor is a genuine self-referential cycle -- it must reuse the
// ancestor's key (closing a real cycle for TopologicalOrder to catch,
// instead of minting a new node and recursing forever) even if it's a v2
// dependency declared SharingMode=false, since sharing intent is moot when
// it's structurally pointing back to itself. A node is added to ancestors
// before its own dependencies are expanded and removed once expansion
// returns, so this only ever matches a still-in-progress ancestor, not an
// already-completed sibling elsewhere in the graph (that case is handled by
// g.sharedByContent instead, for shareable dependencies only).
func (b *GraphBuilder) expandNode(
	ctx context.Context,
	g *Graph,
	key NodeKey,
	bun cnab.ExtendedBundle,
	opts ExplainOpts,
	depth int,
	ancestors map[NodeKey]NodeKey,
) error {
	if depth >= b.maxDepth {
		fmt.Fprintf(b.porter.Err, "warning: dependency graph exceeds max depth of %d, stopping traversal\n", b.maxDepth)
		return nil
	}

	isV2 := bun.HasDependenciesV2()
	if !bun.HasDependenciesV1() && !isV2 {
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

	// v2 resolution iterates a map (ExtendedBundle.ResolveSharedDeps), so
	// locks come back in random order; sort by alias so requires-edges (and
	// therefore inspect's rendered order) are deterministic across runs. v1
	// locks are already ordered by sequence, which sorting would disturb.
	if isV2 {
		sort.Slice(locks, func(i, j int) bool { return locks[i].Alias < locks[j].Alias })
	}

	// Gather each dependency's wiring configuration (v2 only) up front, so
	// dedup keys and wiring edges can both be computed per-lock. Wiring refs
	// are also indexed by target alias here (refsByToAlias) so that, further
	// down, existing-installation matching can determine which of a
	// dependency's own outputs are required by its siblings before deciding
	// whether to reuse an installation for it.
	var v2Requires map[string]v2.Dependency
	var wiringRefs []wiringRef
	var wiringDangling []danglingWiringRef
	var wiringInvalid []invalidWiringRef
	refsByToAlias := make(map[string][]wiringRef)
	if isV2 {
		v2Deps, err := bun.ReadDependenciesV2()
		if err != nil {
			return fmt.Errorf("failed to read v2 dependencies: %w", err)
		}
		v2Requires = v2Deps.Requires

		wiringRefs, wiringDangling, wiringInvalid = extractWiringRefs(v2Requires)
		for _, ref := range wiringRefs {
			refsByToAlias[ref.ToAlias] = append(refsByToAlias[ref.ToAlias], ref)
		}
	}

	// alias -> NodeKey, for resolving wiring refs (which are alias-scoped to
	// this level's Requires map) to actual node keys.
	aliasToKey := make(map[string]NodeKey, len(locks))

	for _, lock := range locks {
		dep, hasDep := v2Requires[lock.Alias]
		var parameters, credentials map[string]string
		if hasDep {
			parameters = dep.Parameters
			credentials = dep.Credentials
		}

		// v1 has no sharing concept and always dedupes by content; v2 only
		// shares an instance when the dependency itself opts in.
		shareable := !isV2 || lock.SharingMode
		ck := contentKey(lock.Reference, parameters, credentials, lock.SharingGroup)

		var childKey NodeKey
		if ancestorKey, onStack := ancestors[ck]; onStack {
			childKey = ancestorKey
		} else if shareable {
			if existing, ok := g.sharedByContent[ck]; ok {
				childKey = existing
			} else {
				childKey = ck
			}
		} else {
			// v2, SharingMode=false, and not a cycle: must never collapse
			// with any other instance, even one with identical content.
			childKey = g.newInstanceKey(ck)
		}
		aliasToKey[lock.Alias] = childKey

		g.addEdge(Edge{
			From:         key,
			To:           childKey,
			Kind:         EdgeKindRequires,
			ToAlias:      lock.Alias,
			SharingMode:  lock.SharingMode,
			SharingGroup: lock.SharingGroup,
		})

		// A node already present means either a shareable dependency fully
		// expanded elsewhere (a diamond) or a cycle back to an ancestor
		// still being expanded (childKey resolved to that ancestor's key,
		// above) -- either way, don't re-pull or re-recurse.
		if _, ok := g.Nodes[childKey]; ok {
			continue
		}

		node := &Node{Key: childKey, Depth: depth}
		g.Nodes[childKey] = node

		// Prefer reusing an already-installed installation over pulling a
		// new instance of the dependency's bundle. A declared bundle
		// interface no longer unconditionally blocks this (#2626): a
		// dependency with no declared interface is always eligible;
		// composeRequiredInterface folds in whatever outputs
		// interface.reference/interface.document add on top of the outputs
		// the parent bundle actually uses -- gated behind
		// FlagDependenciesV2 so a dependency declaring an interface still
		// always falls through to the pull below when the experimental
		// flag is disabled, exactly as it did before #2626.
		//
		// composeRequiredInterface returning an error is handled two ways:
		// errInterfaceReferenceAndDocument (both reference and document
		// set on the interface) is an authoring bug, not a transient
		// failure, so it aborts the node the same way a pull failure does,
		// below. Any other error (e.g. a failed interface.Reference pull)
		// is treated the same as "no match found" rather than aborting the
		// node, so this remains a best-effort optimization on top of the
		// existing pull path.
		if hasDep && (dep.Interface == nil || b.porter.IsFeatureEnabled(experimental.FlagDependenciesV2)) {
			required, err := b.composeRequiredInterface(ctx, lock.Alias, dep, refsByToAlias, opts)
			if err != nil && errors.Is(err, errInterfaceReferenceAndDocument) {
				node.ResolutionFailed = true
				node.ResolutionError = err.Error()
				continue
			}
			if err == nil {
				if inst, err := findExistingInstallation(ctx, b.porter, opts.Namespace, lock, required.Outputs); err == nil && inst != nil {
					node.ResolvedInstallation = inst
					if shareable {
						g.sharedByContent[ck] = childKey
					}
					continue
				}
			}
		}

		childBun, err := b.pullDependencyBundle(ctx, lock.Reference, opts)
		if err != nil {
			node.ResolutionFailed = true
			node.ResolutionError = err.Error()
			continue
		}
		node.Bundle = childBun

		if shareable {
			g.sharedByContent[ck] = childKey
		}

		ancestors[ck] = childKey
		err = b.expandNode(ctx, g, childKey, childBun, opts, depth+1, ancestors)
		delete(ancestors, ck)
		if err != nil {
			return err
		}
	}

	if v2Requires != nil {
		for _, ref := range wiringRefs {
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
		for _, d := range wiringDangling {
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
		for _, inv := range wiringInvalid {
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
		if child.ResolvedInstallation != nil {
			dep.ResolvedInstallation = child.ResolvedInstallation.Namespace + "/" + child.ResolvedInstallation.Name
		}

		if !child.ResolutionFailed && child.ResolvedInstallation == nil {
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
