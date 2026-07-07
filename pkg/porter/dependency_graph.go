package porter

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
)

// NodeKey uniquely identifies a node in a Graph. Two dependency declarations
// that resolve to the same bundle reference, with the same parameters,
// credentials, and sharing group, are considered the same instance and
// collapse to a single node, regardless of the alias(es) used to reach them.
type NodeKey struct {
	// IsRoot is true only for the single node representing the bundle whose
	// graph is being resolved.
	IsRoot bool

	// Reference is the fully version-resolved OCI reference of the bundle,
	// e.g. DependencyLock.Reference.
	Reference string

	// ParametersHash is a stable hash of the dependency's Parameters map.
	ParametersHash string

	// CredentialsHash is a stable hash of the dependency's Credentials map.
	CredentialsHash string

	// SharingGroup is the dependency's sharing group name, if any.
	SharingGroup string
}

func (k NodeKey) String() string {
	if k.IsRoot {
		return "root"
	}
	return k.Reference
}

// EdgeKind distinguishes why an edge exists in the graph.
type EdgeKind string

const (
	// EdgeKindRequires is a structural edge: the From bundle declares To as a dependency.
	EdgeKindRequires EdgeKind = "requires"

	// EdgeKindWiring is a data-flow edge: a parameter, credential, or output
	// value on the From dependency is sourced from an output of the To dependency.
	EdgeKindWiring EdgeKind = "wiring"
)

// WiringDetail describes the specific field that produced a wiring edge.
type WiringDetail struct {
	// Field is the map the reference was found in: "parameters", "credentials", or "outputs".
	Field string

	// FieldName is the key within that map.
	FieldName string

	// SourceOutput is the name of the output on the target dependency being referenced.
	SourceOutput string
}

// Edge is a directed edge in a Graph. From depends on To, meaning To must be
// resolved (and, at execution time, run) before From.
type Edge struct {
	From NodeKey
	To   NodeKey
	Kind EdgeKind

	// ToAlias is the alias used, within the common parent bundle's Requires
	// map, to refer to To. Set for both EdgeKindRequires (the alias the
	// parent declared for this dependency) and EdgeKindWiring (the alias of
	// the sibling dependency being referenced). Stored on the edge itself,
	// rather than on the target Node, because two aliases under the same
	// parent can dedupe to the same node -- only the edge knows which
	// occurrence it represents.
	ToAlias string

	// SharingMode/SharingGroup apply to requires edges, taken from the
	// dependency's own declaration.
	SharingMode  bool
	SharingGroup string

	// Detail is set only when Kind is EdgeKindWiring.
	Detail *WiringDetail
}

// Node is a single resolved dependency (or the root bundle) in a Graph.
type Node struct {
	Key    NodeKey
	Bundle cnab.ExtendedBundle
	Depth  int

	// ResolutionFailed is set when the bundle for this node could not be
	// pulled/resolved; Bundle is the zero value in that case and this node's
	// own dependencies were not expanded.
	ResolutionFailed bool
	ResolutionError  string

	// Warnings holds non-fatal authoring problems found on this node, such
	// as a wiring reference naming a sibling dependency that doesn't exist.
	Warnings []string
}

// Graph is the resolved dependency graph for a bundle: every transitively
// required dependency, deduplicated to one node per unique (reference,
// parameters, credentials, sharing group) instance, plus both structural
// "requires" edges and data-flow "wiring" edges derived from output
// references between sibling dependencies.
type Graph struct {
	Root  NodeKey
	Nodes map[NodeKey]*Node
	Edges []Edge

	// edgesFrom/edgesTo index Edges by From/To so EdgesFrom/EdgesTo are O(1)
	// instead of a linear scan of Edges -- graphToInspectableDependencies
	// calls EdgesFrom once per node occurrence, so an unindexed scan would
	// be O(nodes*edges) over the whole graph.
	edgesFrom map[NodeKey][]Edge
	edgesTo   map[NodeKey][]Edge
}

func newGraph() *Graph {
	return &Graph{
		Nodes:     make(map[NodeKey]*Node),
		edgesFrom: make(map[NodeKey][]Edge),
		edgesTo:   make(map[NodeKey][]Edge),
	}
}

func (g *Graph) addEdge(e Edge) {
	g.Edges = append(g.Edges, e)
	g.edgesFrom[e.From] = append(g.edgesFrom[e.From], e)
	g.edgesTo[e.To] = append(g.edgesTo[e.To], e)
}

// EdgesFrom returns every edge originating at the given node.
func (g *Graph) EdgesFrom(k NodeKey) []Edge {
	return g.edgesFrom[k]
}

// EdgesTo returns every edge terminating at the given node.
func (g *Graph) EdgesTo(k NodeKey) []Edge {
	return g.edgesTo[k]
}

// ErrDependencyCycle is returned when the dependency graph contains a cycle
// (via requires edges, wiring edges, or a combination of both) and therefore
// cannot be executed.
type ErrDependencyCycle struct {
	// Remaining is the set of node references that could not be ordered,
	// i.e. that participate in the cycle.
	Remaining []string
}

func (e ErrDependencyCycle) Error() string {
	return fmt.Sprintf("circular dependency detected involving: %s", strings.Join(e.Remaining, ", "))
}

// TopologicalOrder returns the graph's nodes ordered so that every node
// appears after all of the nodes it depends on (via either requires or
// wiring edges) — i.e. a valid execution/build order, root last. Returns
// ErrDependencyCycle if the combined requires+wiring edge set is not a DAG.
func (g *Graph) TopologicalOrder() ([]*Node, error) {
	// Kahn's algorithm. An edge From->To means From depends on To, so To
	// must be ordered before From: track, for each node, how many
	// not-yet-ordered dependencies it still has (inDegree), and which nodes
	// depend on it (dependents), so we can decrement as each dependency is
	// resolved.
	inDegree := make(map[NodeKey]int, len(g.Nodes))
	dependents := make(map[NodeKey][]NodeKey, len(g.Nodes)) // To -> []From
	for k := range g.Nodes {
		inDegree[k] = 0
	}
	for _, e := range g.Edges {
		inDegree[e.From]++
		dependents[e.To] = append(dependents[e.To], e.From)
	}

	var ready []NodeKey
	for k, d := range inDegree {
		if d == 0 {
			ready = append(ready, k)
		}
	}
	sort.Slice(ready, func(i, j int) bool { return ready[i].String() < ready[j].String() })

	order := make([]*Node, 0, len(g.Nodes))
	for len(ready) > 0 {
		k := ready[0]
		ready = ready[1:]
		order = append(order, g.Nodes[k])

		next := dependents[k]
		sort.Slice(next, func(i, j int) bool { return next[i].String() < next[j].String() })
		for _, from := range next {
			inDegree[from]--
			if inDegree[from] == 0 {
				ready = append(ready, from)
			}
		}
	}

	if len(order) != len(g.Nodes) {
		var remaining []string
		for k, d := range inDegree {
			if d > 0 {
				remaining = append(remaining, k.String())
			}
		}
		sort.Strings(remaining)
		return nil, ErrDependencyCycle{Remaining: remaining}
	}

	return order, nil
}

// computeNodeKey builds the dedup key for a dependency given its resolved
// reference and the wiring configuration that will be applied to it.
func computeNodeKey(reference string, parameters, credentials map[string]string, sharingGroup string) NodeKey {
	return NodeKey{
		Reference:       reference,
		ParametersHash:  hashStringMap(parameters),
		CredentialsHash: hashStringMap(credentials),
		SharingGroup:    sharingGroup,
	}
}

// hashStringMap returns a stable hash of a string map's contents, independent
// of iteration order, for use in dedup keys. Mirrors the pattern used
// elsewhere in this codebase for hashing parameter/credential maps (e.g.
// storage.Run.SetParametersDigest): encoding/json deterministically sorts
// map keys, so hashing the marshaled JSON is enough.
func hashStringMap(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	data, _ := json.Marshal(m)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
