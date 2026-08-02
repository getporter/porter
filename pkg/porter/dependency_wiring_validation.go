package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
)

// validateDependencyWiring builds bun's full dependency graph in strict mode
// (see GraphBuilder.WithStrictWiring) so an unresolvable v2 wiring reference
// (dangling alias, self-reference, or a reference to the root bundle's own
// output) or an unresolvable dependency tree (exceeding the max depth)
// fails fast, before any dependency or the root bundle is executed. Callers
// must only invoke this for bundles with bun.HasDependenciesV2() true; v1
// bundles have no wiring concept and this is a wasted graph traversal for
// them.
func (p *Porter) validateDependencyWiring(ctx context.Context, bun cnab.ExtendedBundle, opts *BundleExecutionOptions) error {
	builder := NewGraphBuilder(p, defaultMaxDependencyDepth).WithStrictWiring()
	_, err := builder.BuildDependencyGraph(ctx, bun, ExplainOpts{
		MaxDependencyDepth:          defaultMaxDependencyDepth,
		DependenciesVersionStrategy: opts.DependenciesVersionStrategy,
		BundleReferenceOptions: BundleReferenceOptions{
			BundlePullOptions: BundlePullOptions{InsecureRegistry: opts.InsecureRegistry},
			installationOptions: installationOptions{
				Namespace: opts.Namespace,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("dependency wiring validation failed: %w", err)
	}
	return nil
}
