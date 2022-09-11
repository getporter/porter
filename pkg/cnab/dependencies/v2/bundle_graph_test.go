package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestEngine_DependOnInstallation(t *testing.T) {
	/*
		A -> B (installation)
		A -> C (bundle)
		c.parameters.connstr <- B.outputs.connstr
	*/

	b := InstallationNode{Key: "b"}
	c := BundleNode{
		Key:      "c",
		Requires: []string{"b"},
	}
	a := BundleNode{
		Key:      "root",
		Requires: []string{"b", "c"},
	}

	g := NewBundleGraph()
	g.RegisterNode(a)
	g.RegisterNode(b)
	g.RegisterNode(c)
	sortedNodes, ok := g.Sort()
	require.True(t, ok, "graph should not be cyclic")

	gotOrder := make([]string, len(sortedNodes))
	for i, node := range sortedNodes {
		gotOrder[i] = node.GetKey()
	}
	wantOrder := []string{
		"b",
		"c",
		"root",
	}
	assert.Equal(t, wantOrder, gotOrder)
}

/*
✅ need to represent new dependency structure on an extended bundle wrapper
(put in cnab-go later)

need to read a bundle and make a BundleGraph
? how to handle a param that isn't a pure assignment, e.g. connstr: ${bundle.deps.VM.outputs.ip}:${bundle.deps.SVC.outputs.port}
? when are templates evaluated as the graph is executed (for simplicity, first draft no composition / templating)

need to resolve dependencies in the graph
* lookup against existing installations
* lookup against semver tags in registry
* lookup against bundle index? when would we look here? (i.e. preferred/registered implementations of interfaces)

need to turn the sorted nodes into an execution plan
execution plan needs:
* bundle to execute and the installation it will become
* parameters and credentials to pass
  * sources:
 	root parameters/creds
	installation outputs

need to write something that can run an execution plan
* knows how to grab sources and pass them into the bundle
*/
