package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildJobRuns(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	installBun := leafTestBundle("mysql")
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/redis:v2.0.0": leafTestBundle("redis"),
	})

	tg := newTestGraph()

	installDep := tg.addNode("localhost:5000/mysql:v1.0.0")
	tg.g.Nodes[installDep] = &Node{Key: installDep, Bundle: installBun}
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
	installations, err := buildJobInstallations(tg.g, order, "dev", storage.NewInstallation("dev", "myapp"))
	require.NoError(t, err)

	actions := map[NodeKey]string{
		tg.g.Root:  cnab.ActionInstall,
		installDep: cnab.ActionInstall,
		skipDep:    JobActionSkip,
		upgradeDep: cnab.ActionUpgrade,
	}

	runs, statuses, err := buildJobRuns(context.Background(), p.Porter, tg.g, jobIDs, installations, actions, ExplainOpts{})
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
	assert.Equal(t, storage.JobStatus{RunID: installRun.ID, Status: cnab.StatusPending}, statuses[installJobID])

	// Upgrade job: Node.Bundle was never pulled by the graph builder, so
	// buildJobRuns must pull it on demand via the mocked registry.
	upgradeJobID := jobIDs[upgradeDep]
	upgradeRun, ok := runs[upgradeJobID]
	require.True(t, ok)
	assert.Equal(t, cnab.ActionUpgrade, upgradeRun.Action)
	assert.Equal(t, "redis", upgradeRun.Bundle.Name)
	assert.Equal(t, storage.JobStatus{RunID: upgradeRun.ID, Status: cnab.StatusPending}, statuses[upgradeJobID])
}
