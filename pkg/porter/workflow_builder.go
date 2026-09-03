package porter

import (
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
)

// buildJobIDs assigns a stable job ID to every node in a topologically
// ordered node list (see Graph.TopologicalOrder), minted in dependency
// order so IDs are deterministic for a given graph traversal.
func buildJobIDs(order []*Node) map[NodeKey]string {
	ids := make(map[NodeKey]string, len(order))
	for _, node := range order {
		ids[node.Key] = cnab.NewULID()
	}
	return ids
}

// buildDependsOn returns the deduped, sorted job IDs of every node
// reachable from key via an outgoing requires or wiring edge -- the
// combined structural and data-flow dependencies a job must wait on.
func buildDependsOn(g *Graph, key NodeKey, jobIDs map[NodeKey]string) []string {
	seen := make(map[string]bool)
	var ids []string
	for _, edge := range g.EdgesFrom(key) {
		id, ok := jobIDs[edge.To]
		if !ok || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// nodeAlias returns the alias a node's Job records: the alias of the first
// (by ToAlias, sorted) requires-edge into it. A node reused by multiple
// parents (a shared/diamond dependency) may have been reached under more
// than one alias; the first sorted alias is used so the result is
// deterministic. Empty for the root node, which has no incoming requires
// edge.
func nodeAlias(g *Graph, key NodeKey) string {
	var aliases []string
	for _, edge := range g.EdgesTo(key) {
		if edge.Kind != EdgeKindRequires {
			continue
		}
		aliases = append(aliases, edge.ToAlias)
	}
	if len(aliases) == 0 {
		return ""
	}
	sort.Strings(aliases)
	return aliases[0]
}

// buildStages groups jobs into storage.Stages using longest-path layering
// over the DependsOn partial order: a job's stage index is one more than
// the highest stage index among the jobs it depends on, or zero if it has
// none. Stages run in series; jobs within a stage have no ordering
// dependency on each other and may run in parallel.
//
// order must be in topological order (dependencies before dependents, as
// returned by Graph.TopologicalOrder) so each job's DependsOn jobs have
// already been assigned a stage index by the time the job itself is
// processed.
func buildStages(g *Graph, order []*Node, jobIDs map[NodeKey]string, dependsOn map[string][]string) []storage.Stage {
	stageIndex := make(map[string]int, len(order))
	maxStage := -1

	for _, node := range order {
		jobID := jobIDs[node.Key]
		idx := 0
		for _, depID := range dependsOn[jobID] {
			if s := stageIndex[depID] + 1; s > idx {
				idx = s
			}
		}
		stageIndex[jobID] = idx
		if idx > maxStage {
			maxStage = idx
		}
	}

	if maxStage < 0 {
		return nil
	}

	stages := make([]storage.Stage, maxStage+1)
	for i := range stages {
		stages[i] = storage.Stage{Jobs: make(map[string]storage.Job)}
	}

	for _, node := range order {
		jobID := jobIDs[node.Key]
		stages[stageIndex[jobID]].Jobs[jobID] = storage.Job{
			ID:           jobID,
			Alias:        nodeAlias(g, node.Key),
			DependsOn:    dependsOn[jobID],
			SharingGroup: node.Key.SharingGroup,
		}
	}

	return stages
}

// buildWorkflowSpec converts a resolved dependency graph into a
// WorkflowSpec: one Job per node -- including a node that turns out to
// already be in sync and needs no bundle action, see JobActionSkip in
// workflow_action_resolver.go -- grouped into Stages by longest-path
// layering over DependsOn.
//
// This only establishes structure: job identity, ordering, and alias.
// Action, Installation, and Credentials are left unset here and are
// populated by ResolveNodeActions, buildJobInstallation, and wireJob*
// respectively (see workflow_action_resolver.go, workflow_installations.go,
// workflow_wiring.go). SchemaType/SchemaVersion/Namespace/MaxParallel/
// DebugMode are also left unset -- the caller is expected to overlay this
// spec onto a storage.NewWorkflow(namespace) value.
//
// Returns an error if any node failed to resolve (Node.ResolutionFailed):
// a workflow can't be generated for a graph that isn't fully resolved,
// regardless of whether the graph was built in strict-wiring mode.
func buildWorkflowSpec(g *Graph) (storage.WorkflowSpec, map[NodeKey]string, error) {
	order, err := g.TopologicalOrder()
	if err != nil {
		return storage.WorkflowSpec{}, nil, err
	}

	for _, node := range order {
		if node.ResolutionFailed {
			return storage.WorkflowSpec{}, nil, fmt.Errorf("cannot generate a workflow: dependency %s failed to resolve: %s", node.Key, node.ResolutionError)
		}
	}

	jobIDs := buildJobIDs(order)

	dependsOn := make(map[string][]string, len(order))
	for _, node := range order {
		dependsOn[jobIDs[node.Key]] = buildDependsOn(g, node.Key, jobIDs)
	}

	root, ok := jobIDs[g.Root]
	if !ok {
		return storage.WorkflowSpec{}, nil, fmt.Errorf("cannot generate a workflow: root node not found in graph")
	}

	spec := storage.WorkflowSpec{
		Root:   root,
		Stages: buildStages(g, order, jobIDs, dependsOn),
	}
	return spec, jobIDs, nil
}
