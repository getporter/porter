package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/storage"
)

// parentInstallationLabel records, on a freshly created dependency
// installation, the root installation whose dependency tree it belongs to.
// Mirrors the "sh.porter.parentInstallation" label runDependencyv2 sets
// (dependencies.go), but always points at the root rather than the
// immediate parent: in a transitive graph a node can be several levels
// deep, and its immediate parent's own identity isn't necessarily decided
// yet at the point a new node is created (see buildJobInstallations).
const parentInstallationLabel = "sh.porter.parentInstallation"

// buildJobInstallations returns the Installation each node's Job should
// carry, keyed by NodeKey:
//   - the root node gets rootInstallation as-is.
//   - a node satisfied by an existing installation (Node.ResolvedInstallation)
//     gets a copy of it.
//   - a brand-new dependency node gets a freshly constructed
//     storage.NewInstallation(namespace, alias), labeled the same way
//     runDependencyv2 labels a new v2 dependency installation
//     (parentInstallationLabel, and sharingGroupLabel when the node
//     belongs to a sharing group).
//
// order must be in topological order (dependencies before dependents, as
// returned by Graph.TopologicalOrder): processing a node in this order
// means every dependency it requires already has its own Installation
// built, so this can immediately record (Installation.AddReference, #2608)
// that the node depends on it, using the node's own just-decided identity.
//
// Purely in-memory -- no Insert*/Update* calls; persisting these records is
// #2647's job.
func buildJobInstallations(g *Graph, order []*Node, namespace string, rootInstallation storage.Installation) (map[NodeKey]storage.Installation, error) {
	built := make(map[NodeKey]*storage.Installation, len(order))

	for _, node := range order {
		key := node.Key

		var inst storage.Installation
		switch {
		case key.IsRoot:
			inst = rootInstallation
		case node.ResolvedInstallation != nil:
			inst = *node.ResolvedInstallation
		default:
			inst = storage.NewInstallation(namespace, nodeAlias(g, key))
			inst.SetLabel(parentInstallationLabel, rootInstallation.String())
			if key.SharingGroup != "" {
				inst.SetLabel(sharingGroupLabel, key.SharingGroup)
			}
		}
		built[key] = &inst

		for _, edge := range g.EdgesFrom(key) {
			if edge.Kind != EdgeKindRequires {
				continue
			}
			child, ok := built[edge.To]
			if !ok {
				return nil, fmt.Errorf("cannot build installation records: dependency %s of %s was not yet built (graph wasn't processed in topological order)", edge.To, key)
			}
			child.AddReference(inst.String(), edge.ToAlias)
		}
	}

	installations := make(map[NodeKey]storage.Installation, len(built))
	for key, inst := range built {
		installations[key] = *inst
	}
	return installations, nil
}
