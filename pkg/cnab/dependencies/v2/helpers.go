package v2

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
)

var _ DependencyResolver = TestResolver{}
var _ BundleGraphResolver = TestResolver{}

// TODO(PEP003): I think we should remove this and just mock the underlying stores used to resolve, e.g. existing installations, or registry queries.
// Otherwise we also have to handle copying values from the dep to the mocked node, or mocking it too
type TestResolver struct {
	Namespace string
	Mocks     map[string]Node
}

func (t TestResolver) ResolveDependency(ctx context.Context, dep Dependency) (Node, bool, error) {
	node, ok := t.Mocks[dep.Key]
	if ok {
		if bunNode, ok := node.(BundleNode); ok {
			bunNode.Parameters = dep.Parameters
			bunNode.Credentials = dep.Credentials
			node = bunNode
		}

		return node, true, nil
	}

	return nil, false, fmt.Errorf("no mock exists for %s", dep.Key)
}

func (t TestResolver) ResolveDependencyGraph(ctx context.Context, bun cnab.ExtendedBundle) (*BundleGraph, error) {
	r := CompositeResolver{
		resolvers: []DependencyResolver{t},
		namespace: t.Namespace,
	}
	return r.ResolveDependencyGraph(ctx, bun)
}
