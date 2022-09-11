package v2

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
)

// DependencyResolver is an interface for various strategies of resolving a
// Dependency to an action in an ExecutionPlan.
type DependencyResolver interface {
	ResolveDependency(ctx context.Context, dep Dependency) (Node, bool, error)
}

type BundleGraphResolver interface {
	ResolveDependencyGraph(ctx context.Context, bun cnab.ExtendedBundle) (*BundleGraph, error)
}
