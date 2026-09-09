package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildJobInstallations_NewDependency(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	dep := NodeKey{Reference: "localhost:5000/mysql:v1.0.0", SharingGroup: "app-db"}
	tg.g.Nodes[dep] = &Node{Key: dep}
	tg.addRequires(tg.g.Root, dep, "db")

	root := storage.NewInstallation("dev", "myapp")

	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)

	installations, err := buildJobInstallations(tg.g, order, "dev", root)
	require.NoError(t, err)

	depInst := installations[dep]
	assert.Equal(t, "dev", depInst.Namespace)
	assert.Equal(t, "db", depInst.Name)
	assert.Equal(t, root.String(), depInst.Labels[parentInstallationLabel])
	assert.Equal(t, "app-db", depInst.Labels[sharingGroupLabel])
	require.Len(t, depInst.Status.References, 1)
	assert.Equal(t, root.String(), depInst.Status.References[0].Installation)
	assert.Equal(t, "db", depInst.Status.References[0].Dependency)

	rootInst := installations[tg.g.Root]
	assert.Equal(t, root.ID, rootInst.ID)
}

func TestBuildJobInstallations_DiamondRecordsBothReferences(t *testing.T) {
	t.Parallel()

	// root requires A and B; both A and B require C (a new dependency).
	tg := newTestGraph()
	depA := tg.addNode("depA-ref")
	depB := tg.addNode("depB-ref")
	depC := tg.addNode("depC-ref")
	tg.addRequires(tg.g.Root, depA, "a")
	tg.addRequires(tg.g.Root, depB, "b")
	tg.addRequires(depA, depC, "c")
	tg.addRequires(depB, depC, "c")

	root := storage.NewInstallation("dev", "myapp")

	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)

	installations, err := buildJobInstallations(tg.g, order, "dev", root)
	require.NoError(t, err)

	cInst := installations[depC]
	require.Len(t, cInst.Status.References, 2)

	aInst := installations[depA]
	bInst := installations[depB]
	referencers := []string{cInst.Status.References[0].Installation, cInst.Status.References[1].Installation}
	assert.ElementsMatch(t, []string{aInst.String(), bInst.String()}, referencers)
}

func TestBuildJobInstallations_ResolvedInstallationIsCopiedAndReferenced(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	dep := tg.addNode("localhost:5000/redis@" + testDigestA)
	existing := storage.NewInstallation("dev", "shared-redis")
	tg.g.Nodes[dep] = &Node{Key: dep, ResolvedInstallation: &existing}
	tg.addRequires(tg.g.Root, dep, "cache")

	root := storage.NewInstallation("dev", "myapp")

	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)

	installations, err := buildJobInstallations(tg.g, order, "dev", root)
	require.NoError(t, err)

	depInst := installations[dep]
	assert.Equal(t, existing.ID, depInst.ID)
	require.Len(t, depInst.Status.References, 1)
	assert.Equal(t, root.String(), depInst.Status.References[0].Installation)
	assert.Equal(t, "cache", depInst.Status.References[0].Dependency)

	// The original existing installation passed in via Node.ResolvedInstallation
	// must not be mutated -- buildJobInstallations works on a copy.
	assert.Empty(t, existing.Status.References)
}

func TestBuildJobInstallations_ResolvedInstallationDoesNotAliasReferences(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	dep := tg.addNode("localhost:5000/redis@" + testDigestA)
	existing := storage.NewInstallation("dev", "shared-redis")

	// backing has spare capacity with a sentinel sitting in the unused
	// slot. A struct-copy-then-append (instead of a defensive copy of
	// References) would silently overwrite that slot via the shared
	// backing array, even though existing.Status.References' own len
	// never changes.
	backing := make([]storage.InstallationReference, 1, 4)
	backing[0] = storage.InstallationReference{Installation: "sentinel-do-not-touch", Dependency: "sentinel"}
	existing.Status.References = backing[:0]

	tg.g.Nodes[dep] = &Node{Key: dep, ResolvedInstallation: &existing}
	tg.addRequires(tg.g.Root, dep, "cache")

	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)

	_, err = buildJobInstallations(tg.g, order, "dev", storage.NewInstallation("dev", "myapp"))
	require.NoError(t, err)

	assert.Equal(t, "sentinel-do-not-touch", backing[:1][0].Installation,
		"shared backing array must not be mutated by AddReference on the copy")
}
