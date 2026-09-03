package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testGraph is a small helper for hand-building a *Graph in tests, without
// going through GraphBuilder.BuildDependencyGraph (which requires pulling
// real bundles). Nodes are added via addNode/addRequires below.
type testGraph struct {
	g *Graph
}

func newTestGraph() *testGraph {
	g := newGraph()
	g.Root = NodeKey{IsRoot: true}
	g.Nodes[g.Root] = &Node{Key: g.Root}
	return &testGraph{g: g}
}

// addNode adds a dependency node identified by reference (used as-is, so
// distinct test nodes must use distinct references).
func (tg *testGraph) addNode(reference string) NodeKey {
	key := NodeKey{Reference: reference}
	tg.g.Nodes[key] = &Node{Key: key}
	return key
}

// addRequires records a structural requires edge: from depends on to,
// reached under the given alias.
func (tg *testGraph) addRequires(from, to NodeKey, alias string) {
	tg.g.addEdge(Edge{From: from, To: to, Kind: EdgeKindRequires, ToAlias: alias})
}

// addWiring records a data-flow wiring edge: from's field sources a value
// from to's output.
func (tg *testGraph) addWiring(from, to NodeKey, toAlias, sourceOutput string) {
	tg.g.addEdge(Edge{
		From: from, To: to, Kind: EdgeKindWiring, ToAlias: toAlias,
		Detail: &WiringDetail{Field: "parameters", FieldName: "p", SourceOutput: sourceOutput},
	})
}

// assertStageContains asserts that stages[index] contains a job with the
// given ID.
func assertStageContains(t *testing.T, stages []storage.Stage, index int, jobID string) {
	t.Helper()
	require.Greater(t, len(stages), index, "expected at least %d stages", index+1)
	_, ok := stages[index].Jobs[jobID]
	assert.True(t, ok, "expected stage %d to contain job %s", index, jobID)
}

// findJob returns the job with the given ID from anywhere in stages,
// failing the test if it isn't found.
func findJob(t *testing.T, stages []storage.Stage, jobID string) storage.Job {
	t.Helper()
	for _, stage := range stages {
		if job, ok := stage.Jobs[jobID]; ok {
			return job
		}
	}
	require.Fail(t, "job not found", "job %s not found in any stage", jobID)
	return storage.Job{}
}

func TestBuildWorkflowSpec_LinearChain(t *testing.T) {
	t.Parallel()

	// root -> depA -> depB
	tg := newTestGraph()
	depA := tg.addNode("depA")
	depB := tg.addNode("depB")
	tg.addRequires(tg.g.Root, depA, "depA")
	tg.addRequires(depA, depB, "depB")

	spec, jobIDs, err := buildWorkflowSpec(tg.g)
	require.NoError(t, err)

	require.Len(t, spec.Stages, 3)
	assertStageContains(t, spec.Stages, 0, jobIDs[depB])
	assertStageContains(t, spec.Stages, 1, jobIDs[depA])
	assertStageContains(t, spec.Stages, 2, jobIDs[tg.g.Root])

	assert.Equal(t, jobIDs[tg.g.Root], spec.Root)

	rootJob := findJob(t, spec.Stages, jobIDs[tg.g.Root])
	assert.Equal(t, []string{jobIDs[depA]}, rootJob.DependsOn)
	assert.Empty(t, rootJob.Alias)

	depAJob := findJob(t, spec.Stages, jobIDs[depA])
	assert.Equal(t, []string{jobIDs[depB]}, depAJob.DependsOn)
	assert.Equal(t, "depA", depAJob.Alias)

	depBJob := findJob(t, spec.Stages, jobIDs[depB])
	assert.Empty(t, depBJob.DependsOn)
	assert.Equal(t, "depB", depBJob.Alias)
}

func TestBuildWorkflowSpec_Diamond(t *testing.T) {
	t.Parallel()

	// root requires A and B; both A and B require C.
	tg := newTestGraph()
	depA := tg.addNode("depA")
	depB := tg.addNode("depB")
	depC := tg.addNode("depC")
	tg.addRequires(tg.g.Root, depA, "a")
	tg.addRequires(tg.g.Root, depB, "b")
	tg.addRequires(depA, depC, "c")
	tg.addRequires(depB, depC, "c")

	spec, jobIDs, err := buildWorkflowSpec(tg.g)
	require.NoError(t, err)

	require.Len(t, spec.Stages, 3)
	assertStageContains(t, spec.Stages, 0, jobIDs[depC])
	assertStageContains(t, spec.Stages, 1, jobIDs[depA])
	assertStageContains(t, spec.Stages, 1, jobIDs[depB])
	assertStageContains(t, spec.Stages, 2, jobIDs[tg.g.Root])

	depCJob := findJob(t, spec.Stages, jobIDs[depC])
	assert.Empty(t, depCJob.DependsOn)
}

func TestBuildWorkflowSpec_IndependentSiblings(t *testing.T) {
	t.Parallel()

	// root requires A and B; A and B are unrelated to each other.
	tg := newTestGraph()
	depA := tg.addNode("depA")
	depB := tg.addNode("depB")
	tg.addRequires(tg.g.Root, depA, "a")
	tg.addRequires(tg.g.Root, depB, "b")

	spec, jobIDs, err := buildWorkflowSpec(tg.g)
	require.NoError(t, err)

	require.Len(t, spec.Stages, 2)
	assertStageContains(t, spec.Stages, 0, jobIDs[depA])
	assertStageContains(t, spec.Stages, 0, jobIDs[depB])
	assertStageContains(t, spec.Stages, 1, jobIDs[tg.g.Root])
}

func TestBuildWorkflowSpec_WiringEdgeCreatesDependency(t *testing.T) {
	t.Parallel()

	// root requires A and B; A's parameter is wired from B's output, even
	// though there's no structural requires edge between them.
	tg := newTestGraph()
	depA := tg.addNode("depA")
	depB := tg.addNode("depB")
	tg.addRequires(tg.g.Root, depA, "a")
	tg.addRequires(tg.g.Root, depB, "b")
	tg.addWiring(depA, depB, "b", "connstr")

	spec, jobIDs, err := buildWorkflowSpec(tg.g)
	require.NoError(t, err)

	require.Len(t, spec.Stages, 3)
	assertStageContains(t, spec.Stages, 0, jobIDs[depB])
	assertStageContains(t, spec.Stages, 1, jobIDs[depA])
	assertStageContains(t, spec.Stages, 2, jobIDs[tg.g.Root])

	depAJob := findJob(t, spec.Stages, jobIDs[depA])
	assert.Equal(t, []string{jobIDs[depB]}, depAJob.DependsOn)
}

func TestBuildWorkflowSpec_ResolutionFailedErrors(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	depA := tg.addNode("depA")
	tg.addRequires(tg.g.Root, depA, "a")
	tg.g.Nodes[depA].ResolutionFailed = true
	tg.g.Nodes[depA].ResolutionError = "pull failed"

	_, _, err := buildWorkflowSpec(tg.g)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pull failed")
}
