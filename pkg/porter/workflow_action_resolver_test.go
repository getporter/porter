package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDigestA = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testDigestB = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func resolvedInstallationNode(key NodeKey, repository, digest string) *Node {
	return &Node{
		Key: key,
		ResolvedInstallation: &storage.Installation{
			InstallationSpec: storage.InstallationSpec{
				Bundle: storage.OCIReferenceParts{Repository: repository, Digest: digest},
			},
			Status: storage.InstallationStatus{BundleDigest: digest},
		},
	}
}

func TestResolveNodeActions(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	newDep := tg.addNode("localhost:5000/new-dep@" + testDigestA)
	inSyncDep := tg.addNode("localhost:5000/mysql@" + testDigestA)
	tg.g.Nodes[inSyncDep] = resolvedInstallationNode(inSyncDep, "localhost:5000/mysql", testDigestA)
	outOfSyncDep := tg.addNode("localhost:5000/redis@" + testDigestA)
	tg.g.Nodes[outOfSyncDep] = resolvedInstallationNode(outOfSyncDep, "localhost:5000/redis", testDigestB)

	actions, err := ResolveNodeActions(tg.g, cnab.ActionUpgrade)
	require.NoError(t, err)

	assert.Equal(t, cnab.ActionUpgrade, actions[tg.g.Root], "root action should pass through unchanged")
	assert.Equal(t, cnab.ActionInstall, actions[newDep], "a node with no resolved installation should be installed")
	assert.Equal(t, JobActionSkip, actions[inSyncDep], "a resolved installation matching the desired digest should be skipped")
	assert.Equal(t, cnab.ActionUpgrade, actions[outOfSyncDep], "a resolved installation with a different digest should be upgraded")
}
