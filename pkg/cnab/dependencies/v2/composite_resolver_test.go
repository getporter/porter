package v2

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompositeResolver_ResolveDependencyGraph(t *testing.T) {
	c := config.NewTestConfig(t)
	c.SetExperimentalFlags(experimental.FlagDependenciesV2)
	c.TestContext.UseFilesystem()
	ctx := context.Background()

	// load our test porter.yaml into a cnab bundle
	m, err := manifest.ReadManifest(c.Context, "testdata/porter.yaml")
	require.NoError(t, err)
	converter := configadapter.NewManifestConverter(c.Config, m, nil, nil)
	bun, err := converter.ToBundle(ctx)
	require.NoError(t, err)

	r := CompositeResolver{
		resolvers: []DependencyResolver{TestResolver{
			Mocks: map[string]Node{
				"root/load-balancer": InstallationNode{Key: "root/load-balancer"},
				"root/mysql":         BundleNode{Key: "root/mysql", Reference: cnab.BundleReference{Definition: cnab.ExtendedBundle{}}},
			}}}}
	g, err := r.ResolveDependencyGraph(ctx, bun)
	require.NoError(t, err)

	sortedNodes, ok := g.Sort()
	require.True(t, ok, "graph should not have a cycle")

	gotOrder := make([]string, len(sortedNodes))
	for i, node := range sortedNodes {
		gotOrder[i] = node.GetKey()
	}
	wantOrder := []string{
		"root/load-balancer",
		"root/mysql",
		"root",
	}
	assert.Equal(t, wantOrder, gotOrder)

	// Check the dependencies of each node
	rootNode, _ := g.GetNode("root")
	require.IsType(t, BundleNode{}, rootNode, "expected the root node to be a bundle")
	require.Equal(t, []string{"root/load-balancer", "root/mysql"}, rootNode.GetRequires(), "expected the root bundle to depend on the load balancer and mysql")

	mysqlNode, _ := g.GetNode("root/mysql")
	require.IsType(t, BundleNode{}, mysqlNode, "expected the mysql node to be a bundle")
	require.Equal(t, []string{"root/load-balancer"}, mysqlNode.GetRequires(), "expected mysql to depend only on the load balancer")

	loadBalancerNode, _ := g.GetNode("root/load-balancer")
	require.IsType(t, InstallationNode{}, loadBalancerNode, "expected the load balancer node to be an installation")
	require.Empty(t, loadBalancerNode.GetRequires(), "expected the load balancer to have no dependencies")
}
