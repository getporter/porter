package v2

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	depsv2ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"get.porter.sh/porter/pkg/storage"
	"github.com/Masterminds/semver/v3"
)

var _ DependencyResolver = CompositeResolver{}
var _ BundleGraphResolver = CompositeResolver{}

// CompositeResolver combines multiple resolution strategies into a single
// resolver that applies each strategy in the proper order to resolve a
// Dependency to an action in an ExecutionPlan.
type CompositeResolver struct {
	namespace string
	resolvers []DependencyResolver
}

func NewCompositeResolver(namespace string, puller BundlePuller, store storage.InstallationProvider) CompositeResolver {
	instResolver := InstallationResolver{
		store:     store,
		namespace: namespace,
	}
	versionResolver := VersionResolver{
		puller: puller,
	}
	return CompositeResolver{
		namespace: namespace,
		resolvers: []DependencyResolver{
			instResolver,
			versionResolver,
			DefaultBundleResolver{puller: puller},
		},
	}
}

func (r CompositeResolver) ResolveDependency(ctx context.Context, dep Dependency) (Node, bool, error) {
	// pull the default bundle if set, and verify that it meets the interface. It's a problem if it doesn't
	// We should stop early if it doesn't work because most likely the interface is defined incorrectly
	// We can check at build time that the bundle will work with all the defaults
	// don't do this at runtime, assume the bundle has been checked

	// build an interface
	// config setting to reuse existing installations

	for _, resolver := range r.resolvers {
		depNode, resolved, err := resolver.ResolveDependency(ctx, dep)
		if err != nil {
			return nil, false, err
		}
		if resolved {
			return depNode, true, nil
		}
	}

	return nil, false, nil
}

func (r CompositeResolver) ResolveDependencyGraph(ctx context.Context, bun cnab.ExtendedBundle) (*BundleGraph, error) {
	g := NewBundleGraph()

	// Add the root bundle
	root := BundleNode{
		Key:       "root",
		Reference: cnab.BundleReference{Definition: bun},
	}

	err := r.addBundleToGraph(ctx, g, root)
	return g, err
}

func (r CompositeResolver) addBundleToGraph(ctx context.Context, g *BundleGraph, node BundleNode) error {
	if _, exists := g.GetNode(node.Key); exists {
		// We have already processed this bundle, return to avoid an infinite loop
		return nil
	}

	// Process dependencies, if it has any
	bun := node.Reference.Definition
	if !bun.HasDependenciesV2() {
		// No deps so let's move on
		g.RegisterNode(node)
		return nil
	}

	deps, err := bun.ReadDependenciesV2()
	if err != nil {
		return err
	}

	node.Requires = make([]string, 0, len(deps.Requires))
	for _, dep := range deps.Requires {
		// Resolve the dependency
		resolved, err := r.resolveDependency(ctx, node.Key, dep)
		if err != nil {
			return err
		}

		// Update the node to track its dependencies
		node.Requires = append(node.Requires, resolved.GetKey())

		//
		// Add the dependency to the graph
		//
		depNode, ok := resolved.(BundleNode)
		if !ok {
			// installations don't have any dependencies so there's nothing left to do
			g.RegisterNode(resolved)
			continue
		}

		// Make connections between the dependency and any outputs of other dependencies that it requires
		requireOutput := func(source depsv2ext.DependencySource) {
			if source.Output == "" {
				return
			}

			outputRequires := node.Key
			if source.Dependency != "" {
				// PEP(003): How do we ensure that these keys are unique in deep graphs where root + current dep key is unique?
				outputRequires = MakeDependencyKey(node.Key, source.Dependency)
			}
			depNode.Requires = append(depNode.Requires, outputRequires)
		}
		for _, source := range dep.Parameters {
			requireOutput(source)
		}
		for _, source := range dep.Credentials {
			requireOutput(source)
		}
		r.addBundleToGraph(ctx, g, depNode)
	}

	g.RegisterNode(node)
	return nil
}

func (r CompositeResolver) resolveDependency(ctx context.Context, parentKey string, dep depsv2ext.Dependency) (Node, error) {
	unresolved := Dependency{
		ParentKey:   parentKey,
		Key:         MakeDependencyKey(parentKey, dep.Name),
		Parameters:  dep.Parameters,
		Credentials: dep.Credentials,
	}
	if dep.Bundle != "" {
		ref, err := cnab.ParseOCIReference(dep.Bundle)
		if err != nil {
			return nil, fmt.Errorf("invalid bundle for dependency %s: %w", unresolved.Key, err)
		}
		unresolved.DefaultBundle = &BundleReferenceSelector{
			Reference: ref,
		}
		if dep.Version != "" {
			unresolved.DefaultBundle.Version, err = semver.NewConstraint(dep.Version)
			if err != nil {
				return nil, err
			}
		}
	}

	if dep.Interface != nil {
		// TODO(PEP003): convert the interface document into a BundleInterfaceSelector
		panic("bundle interfaces are not implemented")
	}

	if dep.Installation != nil {
		unresolved.InstallationSelector = &InstallationSelector{}

		matchNamespaces := make([]string, 0, 2)
		if !dep.Installation.Criteria.IgnoreLabels {
			unresolved.InstallationSelector.Labels = dep.Installation.Labels
		}

		matchNamespaces = append(matchNamespaces, r.namespace)
		if !dep.Installation.Criteria.MatchNamespace && r.namespace != "" {
			// Include the global namespace
			matchNamespaces = append(matchNamespaces, "")
		}
		unresolved.InstallationSelector.Namespaces = matchNamespaces

		if !dep.Installation.Criteria.MatchInterface {
			unresolved.InstallationSelector.Bundle = unresolved.DefaultBundle
		}
	}

	depNode, resolved, err := r.ResolveDependency(ctx, unresolved)
	if err != nil {
		return nil, err
	}

	if !resolved {
		return nil, fmt.Errorf("could  not resolve dependency %s", unresolved.Key)
	}

	return depNode, nil
}
