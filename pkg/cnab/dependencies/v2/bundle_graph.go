package v2

import (
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	depsv2ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"github.com/yourbasic/graph"
)

// BundleGraph is a directed acyclic graph of a bundle and its dependencies
// (which may be other bundles, or installations) It is used to resolve the
// dependency order in which the bundles must be executed.
type BundleGraph struct {
	// nodeKeys is a map from the node key to its index in nodes
	nodeKeys map[string]int
	nodes    []Node
}

func NewBundleGraph() *BundleGraph {
	return &BundleGraph{
		nodeKeys: make(map[string]int),
	}
}

// RegisterNode adds the specified node to the graph
// returning true if the node is already present.
func (g *BundleGraph) RegisterNode(node Node) bool {
	_, exists := g.nodeKeys[node.GetKey()]
	if !exists {
		nodeIndex := len(g.nodes)
		g.nodes = append(g.nodes, node)
		g.nodeKeys[node.GetKey()] = nodeIndex
	}
	return exists
}

func (g *BundleGraph) Sort() ([]Node, bool) {
	dag := graph.New(len(g.nodes))
	for nodeIndex, node := range g.nodes {
		for _, depKey := range node.GetRequires() {
			depIndex, ok := g.nodeKeys[depKey]
			if !ok {
				panic("oops")
			}
			dag.Add(nodeIndex, depIndex)
		}
	}

	indices, ok := graph.TopSort(dag)
	if !ok {
		return nil, false
	}

	// Reverse the sort so that items with no dependencies are listed first
	count := len(indices)
	results := make([]Node, count)
	for i, nodeIndex := range indices {
		results[count-i-1] = g.nodes[nodeIndex]
	}
	return results, true
}

func (g *BundleGraph) GetNode(key string) (Node, bool) {
	if nodeIndex, ok := g.nodeKeys[key]; ok {
		return g.nodes[nodeIndex], true
	}
	return nil, false
}

// Node in a BundleGraph.
type Node interface {
	GetRequires() []string
	GetKey() string
}

var _ Node = BundleNode{}
var _ Node = InstallationNode{}

// BundleNode is a Node in a BundleGraph that represents a dependency on a bundle
// that has not yet been installed.
type BundleNode struct {
	Key       string
	ParentKey string
	Reference cnab.BundleReference
	Requires  []string

	// TODO(PEP003): DO we need to store this? Can we do it somewhere else or hold a reference to the dep and add more to the Node interface?
	Credentials map[string]depsv2ext.DependencySource
	Parameters  map[string]depsv2ext.DependencySource
}

func (d BundleNode) GetKey() string {
	return d.Key
}

func (d BundleNode) GetParentKey() string {
	return d.ParentKey
}

func (d BundleNode) GetRequires() []string {
	sort.Strings(d.Requires)
	return d.Requires
}

func (d BundleNode) IsRoot() bool {
	return d.Key == "root"
}

// InstallationNode is a Node in a BundleGraph that represents a dependency on an
// installed bundle (installation).
type InstallationNode struct {
	Key       string
	ParentKey string
	Namespace string
	Name      string
}

func (d InstallationNode) GetKey() string {
	return d.Key
}

func (d InstallationNode) GetParentKey() string {
	return d.ParentKey
}

func (d InstallationNode) GetRequires() []string {
	return nil
}
