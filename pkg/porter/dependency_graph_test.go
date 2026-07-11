package porter

import (
	"context"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	depsv1 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockPullBundle returns a MockPullBundle func that serves bundle
// definitions from the given map, keyed by the exact OCI reference string.
func newMockPullBundle(bundles map[string]cnab.ExtendedBundle) func(context.Context, cnab.OCIReference, cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
	return func(_ context.Context, ref cnab.OCIReference, _ cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
		bun, ok := bundles[ref.String()]
		if !ok {
			return cnab.BundleReference{}, fmt.Errorf("no mock bundle registered for %s", ref.String())
		}
		return cnab.BundleReference{Reference: ref, Definition: bun}, nil
	}
}

func v1TestBundle(name string, requires map[string]depsv1.Dependency) cnab.ExtendedBundle {
	custom := map[string]any{}
	if len(requires) > 0 {
		custom[cnab.DependenciesV1ExtensionKey] = depsv1.Dependencies{Requires: requires}
	}
	return cnab.ExtendedBundle{Bundle: bundle.Bundle{Name: name, Version: "1.0.0", Custom: custom}}
}

func v2TestBundle(name string, requires map[string]v2.Dependency) cnab.ExtendedBundle {
	custom := map[string]any{}
	if len(requires) > 0 {
		custom[cnab.DependenciesV2ExtensionKey] = v2.Dependencies{Requires: requires}
	}
	return cnab.ExtendedBundle{Bundle: bundle.Bundle{Name: name, Version: "1.0.0", Custom: custom}}
}

func leafTestBundle(name string) cnab.ExtendedBundle {
	return cnab.ExtendedBundle{Bundle: bundle.Bundle{Name: name, Version: "1.0.0"}}
}

func countEdges(edges []Edge, kind EdgeKind) int {
	n := 0
	for _, e := range edges {
		if e.Kind == kind {
			n++
		}
	}
	return n
}

func TestGraphBuilder_NodeDedup(t *testing.T) {
	t.Parallel()

	leaf := leafTestBundle("mysql")

	t.Run("v2 non-shareable dependencies never collapse, even with identical content", func(t *testing.T) {
		t.Parallel()

		// Neither declares a sharing block, so SharingMode defaults to
		// false: per the dependencies v2 extension, "the dependency cannot
		// be shared, even within the same dependency graph" -- so db1 and
		// db2 must each get their own node despite matching content.
		root := v2TestBundle("root", map[string]v2.Dependency{
			"db1": {Bundle: "localhost:5000/mysql:v1.0.0", Parameters: map[string]string{"db-name": "app"}},
			"db2": {Bundle: "localhost:5000/mysql:v1.0.0", Parameters: map[string]string{"db-name": "app"}},
		})

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
			"localhost:5000/mysql:v1.0.0": leaf,
		})

		builder := NewGraphBuilder(p.Porter, 10)
		graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
		require.NoError(t, err)

		// root + two distinct (never-shared) mysql nodes
		assert.Len(t, graph.Nodes, 3)
		assert.Equal(t, 2, countEdges(graph.Edges, EdgeKindRequires))

		deps := graphToInspectableDependencies(graph, graph.Root, 0)
		require.Len(t, deps, 2)
		aliases := []string{deps[0].Alias, deps[1].Alias}
		assert.ElementsMatch(t, []string{"db1", "db2"}, aliases)
		assert.Equal(t, deps[0].Reference, deps[1].Reference)
	})

	t.Run("v2 shareable dependencies with matching group collapse to one node", func(t *testing.T) {
		t.Parallel()

		// Both opt into sharing, with the same group name, so per spec
		// they're the same instance despite being declared under two
		// different aliases.
		shared := v2.SharingCriteria{Mode: true, Group: v2.SharingGroup{Name: "app-db"}}
		root := v2TestBundle("root", map[string]v2.Dependency{
			"db1": {Bundle: "localhost:5000/mysql:v1.0.0", Parameters: map[string]string{"db-name": "app"}, Sharing: shared},
			"db2": {Bundle: "localhost:5000/mysql:v1.0.0", Parameters: map[string]string{"db-name": "app"}, Sharing: shared},
		})

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
			"localhost:5000/mysql:v1.0.0": leaf,
		})

		builder := NewGraphBuilder(p.Porter, 10)
		graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
		require.NoError(t, err)

		// root + one deduped mysql node
		assert.Len(t, graph.Nodes, 2)
		// but two requires-edges, one per alias
		assert.Equal(t, 2, countEdges(graph.Edges, EdgeKindRequires))

		deps := graphToInspectableDependencies(graph, graph.Root, 0)
		require.Len(t, deps, 2)
		aliases := []string{deps[0].Alias, deps[1].Alias}
		assert.ElementsMatch(t, []string{"db1", "db2"}, aliases)
		assert.Equal(t, deps[0].Reference, deps[1].Reference)
	})

	t.Run("v2 shareable dependencies with different groups produce distinct nodes", func(t *testing.T) {
		t.Parallel()

		root := v2TestBundle("root", map[string]v2.Dependency{
			"db1": {Bundle: "localhost:5000/mysql:v1.0.0", Parameters: map[string]string{"db-name": "app"}, Sharing: v2.SharingCriteria{Mode: true, Group: v2.SharingGroup{Name: "group-a"}}},
			"db2": {Bundle: "localhost:5000/mysql:v1.0.0", Parameters: map[string]string{"db-name": "app"}, Sharing: v2.SharingCriteria{Mode: true, Group: v2.SharingGroup{Name: "group-b"}}},
		})

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
			"localhost:5000/mysql:v1.0.0": leaf,
		})

		builder := NewGraphBuilder(p.Porter, 10)
		graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
		require.NoError(t, err)

		// root + two distinct mysql nodes (different sharing groups)
		assert.Len(t, graph.Nodes, 3)
	})

	t.Run("different parameters produce distinct nodes", func(t *testing.T) {
		t.Parallel()

		root := v2TestBundle("root", map[string]v2.Dependency{
			"db1": {Bundle: "localhost:5000/mysql:v1.0.0", Parameters: map[string]string{"db-name": "app1"}},
			"db2": {Bundle: "localhost:5000/mysql:v1.0.0", Parameters: map[string]string{"db-name": "app2"}},
		})

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
			"localhost:5000/mysql:v1.0.0": leaf,
		})

		builder := NewGraphBuilder(p.Porter, 10)
		graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
		require.NoError(t, err)

		// root + two distinct mysql nodes (different parameters)
		assert.Len(t, graph.Nodes, 3)
	})

	t.Run("v1 dependencies always dedupe by content (no sharing concept)", func(t *testing.T) {
		t.Parallel()

		root := v1TestBundle("root", map[string]depsv1.Dependency{
			"db1": {Bundle: "localhost:5000/mysql:v1.0.0"},
			"db2": {Bundle: "localhost:5000/mysql:v1.0.0"},
		})

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
			"localhost:5000/mysql:v1.0.0": leaf,
		})

		builder := NewGraphBuilder(p.Porter, 10)
		graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
		require.NoError(t, err)

		// root + one deduped mysql node
		assert.Len(t, graph.Nodes, 2)
		assert.Equal(t, 2, countEdges(graph.Edges, EdgeKindRequires))
	})
}

func TestGraphBuilder_WiringEdges(t *testing.T) {
	t.Parallel()

	// Mirrors the real sibling-wiring example in
	// pkg/cnab/config-adapter/testdata/myenv-depsv2.bundle.json
	root := v2TestBundle("root", map[string]v2.Dependency{
		"app": {
			Bundle: "localhost:5000/myapp:v1.2.3",
			Credentials: map[string]string{
				"db-connstr": "${bundle.dependencies.infra.outputs.mysql-connstr}",
			},
		},
		"infra": {
			Bundle: "localhost:5000/myinfra:v0.1.0",
		},
	})

	p := NewTestPorter(t)
	defer p.Close()
	leaf := leafTestBundle("leaf")
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/myapp:v1.2.3":   leaf,
		"localhost:5000/myinfra:v0.1.0": leaf,
	})

	builder := NewGraphBuilder(p.Porter, 10)
	graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.NoError(t, err)

	require.Equal(t, 1, countEdges(graph.Edges, EdgeKindWiring))
	for _, e := range graph.Edges {
		if e.Kind != EdgeKindWiring {
			continue
		}
		require.NotNil(t, e.Detail)
		assert.Equal(t, "credentials", e.Detail.Field)
		assert.Equal(t, "db-connstr", e.Detail.FieldName)
		assert.Equal(t, "mysql-connstr", e.Detail.SourceOutput)
	}

	deps := graphToInspectableDependencies(graph, graph.Root, 0)
	var appDep *InspectableDependency
	for i := range deps {
		if deps[i].Alias == "app" {
			appDep = &deps[i]
		}
	}
	require.NotNil(t, appDep)
	require.Len(t, appDep.WiringEdges, 1)
	assert.Equal(t, "infra", appDep.WiringEdges[0].SourceDependencyAlias)
	assert.Equal(t, "mysql-connstr", appDep.WiringEdges[0].SourceOutput)
}

func TestGraphBuilder_V1BundlesNeverProduceWiringEdges(t *testing.T) {
	t.Parallel()

	root := v1TestBundle("root", map[string]depsv1.Dependency{
		"mysql": {Bundle: "localhost:5000/mysql:v1.0.0"},
		"nginx": {Bundle: "localhost:5000/nginx:v1.0.0"},
	})

	p := NewTestPorter(t)
	defer p.Close()
	leaf := leafTestBundle("leaf")
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/mysql:v1.0.0": leaf,
		"localhost:5000/nginx:v1.0.0": leaf,
	})

	builder := NewGraphBuilder(p.Porter, 10)
	graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.NoError(t, err)

	assert.Equal(t, 0, countEdges(graph.Edges, EdgeKindWiring))
}

func TestGraphBuilder_CycleDetection_Requires(t *testing.T) {
	t.Parallel()

	aBun := v1TestBundle("a", map[string]depsv1.Dependency{
		"b": {Bundle: "localhost:5000/b:v1.0.0"},
	})
	bBun := v1TestBundle("b", map[string]depsv1.Dependency{
		"a": {Bundle: "localhost:5000/a:v1.0.0"},
	})
	root := v1TestBundle("root", map[string]depsv1.Dependency{
		"a": {Bundle: "localhost:5000/a:v1.0.0"},
	})

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/a:v1.0.0": aBun,
		"localhost:5000/b:v1.0.0": bBun,
	})

	builder := NewGraphBuilder(p.Porter, 10)
	_, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.Error(t, err)
	var cycleErr ErrDependencyCycle
	require.ErrorAs(t, err, &cycleErr)
	// Both cycle participants should be reported (not just the one node
	// where the cycle happened to be detected while pulling), plus "root"
	// itself, which also can never be ordered since it transitively depends
	// on the unresolvable a/b pair.
	assert.ElementsMatch(t, []string{"localhost:5000/a:v1.0.0", "localhost:5000/b:v1.0.0", "root"}, cycleErr.Remaining)
}

// TestGraphBuilder_CycleDetection_NonShareableSelfReference covers a mutual
// requires cycle built entirely from v2 dependencies with SharingMode=false
// (the default): a and b requiring each other must never collapse onto a
// shared node by content (see TestGraphBuilder_NodeDedup), yet the recursion
// still needs to recognize b's "a" as the *same* content already being
// expanded by an in-progress ancestor -- otherwise this would recurse
// forever (bounded only by silently hitting maxDepth) instead of reporting
// ErrDependencyCycle.
func TestGraphBuilder_CycleDetection_NonShareableSelfReference(t *testing.T) {
	t.Parallel()

	aBun := v2TestBundle("a", map[string]v2.Dependency{
		"b": {Bundle: "localhost:5000/b:v1.0.0"},
	})
	bBun := v2TestBundle("b", map[string]v2.Dependency{
		"a": {Bundle: "localhost:5000/a:v1.0.0"},
	})
	root := v2TestBundle("root", map[string]v2.Dependency{
		"a": {Bundle: "localhost:5000/a:v1.0.0"},
	})

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/a:v1.0.0": aBun,
		"localhost:5000/b:v1.0.0": bBun,
	})

	builder := NewGraphBuilder(p.Porter, 10)
	_, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.Error(t, err)
	var cycleErr ErrDependencyCycle
	require.ErrorAs(t, err, &cycleErr)
	assert.ElementsMatch(t, []string{"localhost:5000/a:v1.0.0", "localhost:5000/b:v1.0.0", "root"}, cycleErr.Remaining)
}

// TestGraphBuilder_SelfReferentialWiringIsAWarningNotACycle covers a
// dependency's own field referencing its own alias (a plausible copy-paste
// typo) -- this must not be treated as a valid wiring edge (which would
// create an unresolvable self-loop and hard-fail the whole graph), but as a
// scoped warning on that node, same as any other invalid wiring reference.
func TestGraphBuilder_SelfReferentialWiringIsAWarningNotACycle(t *testing.T) {
	t.Parallel()

	root := v2TestBundle("root", map[string]v2.Dependency{
		"app": {
			Bundle: "localhost:5000/myapp:v1.0.0",
			Credentials: map[string]string{
				"conn": "${bundle.dependencies.app.outputs.foo}",
			},
		},
	})

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/myapp:v1.0.0": leafTestBundle("app"),
	})

	builder := NewGraphBuilder(p.Porter, 10)
	graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.NoError(t, err, "a self-referential wiring value should be a warning, not a cycle failure")

	assert.Equal(t, 0, countEdges(graph.Edges, EdgeKindWiring))

	deps := graphToInspectableDependencies(graph, graph.Root, 0)
	require.Len(t, deps, 1)
	require.Len(t, deps[0].Warnings, 1)
	assert.Contains(t, deps[0].Warnings[0], "app")
	assert.Contains(t, deps[0].Warnings[0], "its own output")
}

// TestGraphBuilder_WiringEdgesAreDeterministicallyOrdered guards against the
// wiring-edge order flapping between runs due to Go's randomized map
// iteration -- extractWiringRefs must sort its output.
func TestGraphBuilder_WiringEdgesAreDeterministicallyOrdered(t *testing.T) {
	t.Parallel()

	root := v2TestBundle("root", map[string]v2.Dependency{
		"app": {
			Bundle: "localhost:5000/myapp:v1.0.0",
			Credentials: map[string]string{
				"db-connstr": "${bundle.dependencies.infra.outputs.mysql-connstr}",
			},
			Outputs: map[string]string{
				"endpoint": "${bundle.dependencies.infra.outputs.ip}",
			},
		},
		"infra": {
			Bundle: "localhost:5000/myinfra:v0.1.0",
		},
	})

	leaf := leafTestBundle("leaf")

	var first []WiringEdgeSummary
	for range 20 {
		p := NewTestPorter(t)
		p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
			"localhost:5000/myapp:v1.0.0":   leaf,
			"localhost:5000/myinfra:v0.1.0": leaf,
		})

		builder := NewGraphBuilder(p.Porter, 10)
		graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
		require.NoError(t, err)
		p.Close()

		deps := graphToInspectableDependencies(graph, graph.Root, 0)
		var appDep *InspectableDependency
		for j := range deps {
			if deps[j].Alias == "app" {
				appDep = &deps[j]
			}
		}
		require.NotNil(t, appDep)
		require.Len(t, appDep.WiringEdges, 2)

		if first == nil {
			first = appDep.WiringEdges
			continue
		}
		assert.Equal(t, first, appDep.WiringEdges, "wiring edge order must be deterministic across runs")
	}
}

// TestGraphBuilder_CycleDetection_Wiring constructs a cycle that only exists
// once requires-edges and wiring-edges are combined: under root's dependency
// "x", siblings "y" and "z" are declared; y's credential is wired from z's
// output (edge y->z), and z's own bundle also depends on y with an
// identical definition so it dedupes to the same node (edge z->y). Neither
// edge alone is a cycle; together, y->z->y is.
func TestGraphBuilder_CycleDetection_Wiring(t *testing.T) {
	t.Parallel()

	// Mode=true with a matching group makes x's and z's "y" entries the
	// same shareable instance, so z's own requirement of "y" collapses back
	// onto the "y" node reached from x -- that's what manufactures the
	// cycle below. Without an explicit shared group, SharingMode defaults
	// to false and the two "y" declarations would correctly stay separate,
	// non-shareable instances (see TestGraphBuilder_NodeDedup), and there
	// would be no cycle at all.
	yEntry := v2.Dependency{
		Bundle:      "localhost:5000/y:v1.0.0",
		Credentials: map[string]string{"conn": "${bundle.dependencies.z.outputs.out}"},
		Sharing:     v2.SharingCriteria{Mode: true, Group: v2.SharingGroup{Name: "y-shared"}},
	}

	xBun := v2TestBundle("x", map[string]v2.Dependency{
		"y": yEntry,
		"z": {Bundle: "localhost:5000/z:v1.0.0"},
	})
	zBun := v2TestBundle("z", map[string]v2.Dependency{
		// Same content as x's "y" entry, so it dedupes to the same node.
		"y": yEntry,
	})
	root := v2TestBundle("root", map[string]v2.Dependency{
		"x": {Bundle: "localhost:5000/x:v1.0.0"},
	})

	p := NewTestPorter(t)
	defer p.Close()
	leaf := leafTestBundle("y")
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/x:v1.0.0": xBun,
		"localhost:5000/z:v1.0.0": zBun,
		"localhost:5000/y:v1.0.0": leaf,
	})

	builder := NewGraphBuilder(p.Porter, 10)
	_, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.Error(t, err)
	var cycleErr ErrDependencyCycle
	require.ErrorAs(t, err, &cycleErr)
}

func TestGraphBuilder_LegitimateDiamondIsNotACycle(t *testing.T) {
	t.Parallel()

	cBun := leafTestBundle("c")
	aBun := v1TestBundle("a", map[string]depsv1.Dependency{
		"c": {Bundle: "localhost:5000/c:v1.0.0"},
	})
	bBun := v1TestBundle("b", map[string]depsv1.Dependency{
		"c": {Bundle: "localhost:5000/c:v1.0.0"},
	})
	root := v1TestBundle("root", map[string]depsv1.Dependency{
		"a": {Bundle: "localhost:5000/a:v1.0.0"},
		"b": {Bundle: "localhost:5000/b:v1.0.0"},
	})

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/a:v1.0.0": aBun,
		"localhost:5000/b:v1.0.0": bBun,
		"localhost:5000/c:v1.0.0": cBun,
	})

	builder := NewGraphBuilder(p.Porter, 10)
	graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.NoError(t, err)

	// root, a, b, c -- c is deduped, not duplicated
	assert.Len(t, graph.Nodes, 4)

	order, err := graph.TopologicalOrder()
	require.NoError(t, err)

	indexOf := func(ref string) int {
		for i, n := range order {
			if n.Key.Reference == ref {
				return i
			}
		}
		t.Fatalf("node %s not found in topological order", ref)
		return -1
	}

	cIdx := indexOf("localhost:5000/c:v1.0.0")
	aIdx := indexOf("localhost:5000/a:v1.0.0")
	bIdx := indexOf("localhost:5000/b:v1.0.0")
	assert.Less(t, cIdx, aIdx, "c must be ordered before a")
	assert.Less(t, cIdx, bIdx, "c must be ordered before b")

	// The root node should be last: it (transitively) depends on everything.
	assert.True(t, order[len(order)-1].Key.IsRoot, "root should be resolved last")
}

func TestGraphBuilder_MaxDepth(t *testing.T) {
	t.Parallel()

	bBun := v1TestBundle("b", map[string]depsv1.Dependency{
		"c": {Bundle: "localhost:5000/c:v1.0.0"},
	})
	aBun := v1TestBundle("a", map[string]depsv1.Dependency{
		"b": {Bundle: "localhost:5000/b:v1.0.0"},
	})
	root := v1TestBundle("root", map[string]depsv1.Dependency{
		"a": {Bundle: "localhost:5000/a:v1.0.0"},
	})

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/a:v1.0.0": aBun,
		"localhost:5000/b:v1.0.0": bBun,
		"localhost:5000/c:v1.0.0": leafTestBundle("c"),
	})

	// maxDepth=2 allows depth-0 (a) and depth-1 (b) to be labeled, but must
	// not expand b's own dependencies (which would be depth-2).
	builder := NewGraphBuilder(p.Porter, 2)
	graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 2})
	require.NoError(t, err)

	deps := graphToInspectableDependencies(graph, graph.Root, 0)
	require.Len(t, deps, 1)
	assert.Equal(t, "a", deps[0].Alias)
	assert.Equal(t, 0, deps[0].Depth)
	require.Len(t, deps[0].Dependencies, 1)
	assert.Equal(t, "b", deps[0].Dependencies[0].Alias)
	assert.Equal(t, 1, deps[0].Dependencies[0].Depth)
	assert.Len(t, deps[0].Dependencies[0].Dependencies, 0, "should not traverse beyond max depth")
}

func TestGraphBuilder_SoftResolutionFailurePreserved(t *testing.T) {
	t.Parallel()

	root := v1TestBundle("root", map[string]depsv1.Dependency{
		"missing": {Bundle: "localhost:5000/missing:v1.0.0"},
	})

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		// intentionally empty: "missing" has no registered mock bundle
	})

	builder := NewGraphBuilder(p.Porter, 10)
	graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.NoError(t, err, "a single dependency's pull failure should not abort the whole graph build")

	deps := graphToInspectableDependencies(graph, graph.Root, 0)
	require.Len(t, deps, 1)
	assert.True(t, deps[0].ResolutionFailed)
	assert.NotEmpty(t, deps[0].ResolutionError)
	assert.Len(t, deps[0].Dependencies, 0)

	_, err = graph.TopologicalOrder()
	assert.NoError(t, err, "a failed node with no outgoing edges should not block topological ordering")
}

func TestGraphBuilder_DanglingWiringReference(t *testing.T) {
	t.Parallel()

	root := v2TestBundle("root", map[string]v2.Dependency{
		"app": {
			Bundle: "localhost:5000/myapp:v1.0.0",
			Credentials: map[string]string{
				"conn": "${bundle.dependencies.doesnotexist.outputs.foo}",
			},
		},
	})

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/myapp:v1.0.0": leafTestBundle("app"),
	})

	builder := NewGraphBuilder(p.Porter, 10)
	graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.NoError(t, err, "a dangling wiring reference should be a warning, not a hard failure")

	assert.Equal(t, 0, countEdges(graph.Edges, EdgeKindWiring))

	deps := graphToInspectableDependencies(graph, graph.Root, 0)
	require.Len(t, deps, 1)
	require.Len(t, deps[0].Warnings, 1)
	assert.Contains(t, deps[0].Warnings[0], "doesnotexist")
}

// TestGraphBuilder_RootOutputReferenceIsAWarning covers a dependency's field
// referencing the ROOT bundle's own output via the long-form syntax
// (`${bundle.outputs.X}`), an easy typo for the valid short-form
// (`${outputs.X}`, meaning the dependency's own output). This can never be
// resolved (dependencies run before the root bundle produces outputs), so it
// must be surfaced as a warning rather than silently vanishing.
func TestGraphBuilder_RootOutputReferenceIsAWarning(t *testing.T) {
	t.Parallel()

	root := v2TestBundle("root", map[string]v2.Dependency{
		"app": {
			Bundle: "localhost:5000/myapp:v1.0.0",
			Parameters: map[string]string{
				"x": "${bundle.outputs.foo}",
			},
		},
	})

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/myapp:v1.0.0": leafTestBundle("app"),
	})

	builder := NewGraphBuilder(p.Porter, 10)
	graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.NoError(t, err, "a root-output reference should be a warning, not a hard failure")

	assert.Equal(t, 0, countEdges(graph.Edges, EdgeKindWiring))

	deps := graphToInspectableDependencies(graph, graph.Root, 0)
	require.Len(t, deps, 1)
	require.Len(t, deps[0].Warnings, 1)
	assert.Contains(t, deps[0].Warnings[0], "root bundle's own output")
}

// TestGraphBuilder_DiamondAtDifferentDepthsIsRelabeledCorrectly reuses a
// cached (memoized) subtree for the same shared dependency reached at two
// DIFFERENT depths, and checks that the depth-relabeling on reuse is
// correct all the way down into the shared node's own nested dependencies,
// not just the shared node itself.
func TestGraphBuilder_DiamondAtDifferentDepthsIsRelabeledCorrectly(t *testing.T) {
	t.Parallel()

	dBun := leafTestBundle("d")
	cBun := v1TestBundle("c", map[string]depsv1.Dependency{
		"d": {Bundle: "localhost:5000/d:v1.0.0"},
	})
	bBun := v1TestBundle("b", map[string]depsv1.Dependency{
		"c": {Bundle: "localhost:5000/c:v1.0.0"},
	})
	aBun := v1TestBundle("a", map[string]depsv1.Dependency{
		"b": {Bundle: "localhost:5000/b:v1.0.0"},
	})
	eBun := v1TestBundle("e", map[string]depsv1.Dependency{
		"c": {Bundle: "localhost:5000/c:v1.0.0"},
	})
	root := v1TestBundle("root", map[string]depsv1.Dependency{
		"a": {Bundle: "localhost:5000/a:v1.0.0"}, // root -> a -> b -> c -> d (c at depth 2, d at depth 3)
		"e": {Bundle: "localhost:5000/e:v1.0.0"}, // root -> e -> c -> d (c at depth 1, d at depth 2)
	})

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/a:v1.0.0": aBun,
		"localhost:5000/b:v1.0.0": bBun,
		"localhost:5000/c:v1.0.0": cBun,
		"localhost:5000/d:v1.0.0": dBun,
		"localhost:5000/e:v1.0.0": eBun,
	})

	builder := NewGraphBuilder(p.Porter, 10)
	graph, err := builder.BuildDependencyGraph(context.Background(), root, ExplainOpts{MaxDependencyDepth: 10})
	require.NoError(t, err)

	// c is deduped to a single node despite being reached at two depths.
	assert.Len(t, graph.Nodes, 6) // root, a, b, c, d, e

	deps := graphToInspectableDependencies(graph, graph.Root, 0)
	require.Len(t, deps, 2)

	var aDep, eDep *InspectableDependency
	for i := range deps {
		switch deps[i].Alias {
		case "a":
			aDep = &deps[i]
		case "e":
			eDep = &deps[i]
		}
	}
	require.NotNil(t, aDep)
	require.NotNil(t, eDep)

	// root -> a(0) -> b(1) -> c(2) -> d(3)
	bDep := aDep.Dependencies[0]
	assert.Equal(t, "b", bDep.Alias)
	assert.Equal(t, 1, bDep.Depth)
	cUnderB := bDep.Dependencies[0]
	assert.Equal(t, "c", cUnderB.Alias)
	assert.Equal(t, 2, cUnderB.Depth)
	dUnderB := cUnderB.Dependencies[0]
	assert.Equal(t, "d", dUnderB.Alias)
	assert.Equal(t, 3, dUnderB.Depth)

	// root -> e(0) -> c(1) -> d(2)
	cUnderE := eDep.Dependencies[0]
	assert.Equal(t, "c", cUnderE.Alias)
	assert.Equal(t, 1, cUnderE.Depth)
	dUnderE := cUnderE.Dependencies[0]
	assert.Equal(t, "d", dUnderE.Alias)
	assert.Equal(t, 2, dUnderE.Depth)
}
