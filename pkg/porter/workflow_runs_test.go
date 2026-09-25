package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/storage"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildJobRuns(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	installBun := leafTestBundle("mysql")
	pullDigests := map[string]string{
		"localhost:5000/redis:v2.0.0": testDigestB,
	}
	mockPull := newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/redis:v2.0.0": leafTestBundle("redis"),
	})
	p.TestRegistry.MockPullBundle = func(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
		bunRef, err := mockPull(ctx, ref, opts)
		bunRef.Digest = digest.Digest(pullDigests[ref.String()])
		return bunRef, err
	}

	tg := newTestGraph()

	installDep := tg.addNode("localhost:5000/mysql:v1.0.0")
	tg.g.Nodes[installDep] = &Node{Key: installDep, Bundle: installBun, Digest: testDigestA}
	tg.addRequires(tg.g.Root, installDep, "db")

	skipDep := tg.addNode("localhost:5000/cache@" + testDigestA)
	skipInst := storage.NewInstallation("dev", "cache")
	skipInst.Bundle = storage.OCIReferenceParts{Repository: "localhost:5000/cache", Digest: testDigestA}
	skipInst.Status.BundleDigest = testDigestA
	tg.g.Nodes[skipDep] = &Node{Key: skipDep, ResolvedInstallation: &skipInst}
	tg.addRequires(tg.g.Root, skipDep, "cache")

	upgradeDep := NodeKey{Reference: "localhost:5000/redis:v2.0.0"}
	upgradeInst := storage.NewInstallation("dev", "redis")
	upgradeInst.Bundle = storage.OCIReferenceParts{Repository: "localhost:5000/redis", Tag: "v1.0.0"}
	upgradeInst.Status.BundleDigest = testDigestB
	tg.g.Nodes[upgradeDep] = &Node{Key: upgradeDep, ResolvedInstallation: &upgradeInst}
	tg.addRequires(tg.g.Root, upgradeDep, "redis")

	tg.g.Nodes[tg.g.Root].Bundle = leafTestBundle("root")

	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)

	jobIDs := buildJobIDs(order)
	rootInst := storage.NewInstallation("dev", "myapp")
	// Stale tracked reference/digest: must not leak into the root run.
	rootInst.Bundle = storage.OCIReferenceParts{Repository: "localhost:5000/myapp", Digest: testDigestB}
	rootRef := cnab.BundleReference{
		Reference: cnab.MustParseOCIReference("localhost:5000/myapp:v2.0.0"),
		Digest:    digest.Digest(testDigestA),
	}
	installations, err := buildJobInstallations(tg.g, order, "dev", rootInst)
	require.NoError(t, err)

	actions := map[NodeKey]string{
		tg.g.Root:  cnab.ActionInstall,
		installDep: cnab.ActionInstall,
		skipDep:    JobActionSkip,
		upgradeDep: cnab.ActionUpgrade,
	}

	runs, statuses, err := buildJobRuns(context.Background(), p.Porter, tg.g, jobIDs, installations, actions, rootRef, ExplainOpts{})
	require.NoError(t, err)

	// Skipped job: no Run, status is immediately succeeded.
	skipJobID := jobIDs[skipDep]
	_, hasRun := runs[skipJobID]
	assert.False(t, hasRun)
	assert.Equal(t, storage.JobStatus{Status: cnab.StatusSucceeded}, statuses[skipJobID])

	// Install job: Run created from the already-pulled Node.Bundle.
	installJobID := jobIDs[installDep]
	installRun, ok := runs[installJobID]
	require.True(t, ok)
	assert.Equal(t, cnab.ActionInstall, installRun.Action)
	assert.Equal(t, "localhost:5000/mysql:v1.0.0", installRun.BundleReference)
	assert.Equal(t, "mysql", installRun.Bundle.Name)
	// Tag-only reference: digest is the one recorded on the node by the graph pull.
	assert.Equal(t, testDigestA, installRun.BundleDigest)

	rootRun, ok := runs[jobIDs[tg.g.Root]]
	require.True(t, ok)
	assert.Equal(t, testDigestA, rootRun.BundleDigest)
	assert.Equal(t, "localhost:5000/myapp:v2.0.0", rootRun.BundleReference)
	assert.Equal(t, storage.JobStatus{RunID: installRun.ID, Status: cnab.StatusPending}, statuses[installJobID])

	// Upgrade job: Node.Bundle was never pulled by the graph builder, so
	// buildJobRuns must pull it on demand via the mocked registry.
	upgradeJobID := jobIDs[upgradeDep]
	upgradeRun, ok := runs[upgradeJobID]
	require.True(t, ok)
	assert.Equal(t, cnab.ActionUpgrade, upgradeRun.Action)
	assert.Equal(t, "redis", upgradeRun.Bundle.Name)
	assert.Equal(t, testDigestB, upgradeRun.BundleDigest)
	assert.Equal(t, storage.JobStatus{RunID: upgradeRun.ID, Status: cnab.StatusPending}, statuses[upgradeJobID])
}

func TestBuildJobRuns_UsesNodeDigestWithoutRepulling(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = func(context.Context, cnab.OCIReference, cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
		return cnab.BundleReference{}, assert.AnError
	}

	const ref = "localhost:5000/mysql:v1.0.0"
	tg := newTestGraph()
	dep := tg.addNode(ref)
	// Two nodes could share a moving tag yet hold different digests; each
	// run must use its own node's digest, never re-resolve the tag.
	tg.g.Nodes[dep] = &Node{Key: dep, Bundle: leafTestBundle("mysql"), Digest: testDigestB}
	tg.addRequires(tg.g.Root, dep, "db")
	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)
	jobIDs := buildJobIDs(order)
	installations, err := buildJobInstallations(tg.g, order, "dev", storage.NewInstallation("dev", "myapp"))
	require.NoError(t, err)

	rootRef := cnab.BundleReference{Reference: cnab.MustParseOCIReference("localhost:5000/myapp:v1.0.0"), Digest: digest.Digest(testDigestA)}
	actions := map[NodeKey]string{tg.g.Root: cnab.ActionInstall, dep: cnab.ActionInstall}

	// Force must have no effect, and no registry call may be made.
	opts := ExplainOpts{BundleReferenceOptions: BundleReferenceOptions{BundlePullOptions: BundlePullOptions{Force: true}}}
	runs, _, err := buildJobRuns(context.Background(), p.Porter, tg.g, jobIDs, installations, actions, rootRef, opts)
	require.NoError(t, err)
	assert.Equal(t, testDigestB, runs[jobIDs[dep]].BundleDigest)
}
