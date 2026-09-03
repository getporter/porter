package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
)

// pullBundleForNode pulls the bundle for a graph node on demand. Mirrors
// GraphBuilder.pullDependencyBundle (dependency_graph_builder.go), used
// here instead of that method because it's only needed for a node
// GraphBuilder itself never pulled -- see buildJobRuns.
func pullBundleForNode(ctx context.Context, p *Porter, ref string, opts ExplainOpts) (cnab.ExtendedBundle, error) {
	pullOpts := BundlePullOptions{
		Reference:        ref,
		InsecureRegistry: opts.InsecureRegistry,
		Force:            opts.Force,
	}

	cachedBundle, err := p.PullBundle(ctx, pullOpts)
	if err != nil {
		return cnab.ExtendedBundle{}, fmt.Errorf("failed to pull bundle %s: %w", ref, err)
	}

	return cachedBundle.Definition, nil
}

// bundleReferenceString returns the bundle reference to record on a job's
// Run: key.Reference for a dependency node (always set, once resolved), or
// -- for the root node, whose NodeKey carries no reference -- the
// installation's own recorded bundle reference.
func bundleReferenceString(key NodeKey, inst storage.Installation) string {
	if key.Reference != "" {
		return key.Reference
	}
	if ref, ok, err := inst.Bundle.GetBundleReference(); err == nil && ok {
		return ref.String()
	}
	return ""
}

// buildJobRuns creates a Pending storage.Run for every job whose action
// requires one (i.e. not JobActionSkip), and a WorkflowStatus.JobStatus
// entry for every job in the graph, including skipped ones (marked
// succeeded immediately, since they require no bundle action).
//
// installations and actions must be the results of buildJobInstallations
// and ResolveNodeActions for the same graph/jobIDs.
//
// For a brand-new dependency node, Node.Bundle is already populated (it
// was pulled while building the graph). For an existing installation being
// upgraded (Node.ResolvedInstallation != nil), GraphBuilder never pulled a
// bundle -- it short-circuited via findExistingInstallation -- so this
// pulls one on demand via pullBundleForNode.
//
// Purely in-memory -- no Insert*/Update* calls; persisting these records is
// #2647's job. "Pending" isn't a field on Run (it has none); it's carried
// entirely by the returned JobStatus.
func buildJobRuns(
	ctx context.Context,
	p *Porter,
	g *Graph,
	jobIDs map[NodeKey]string,
	installations map[NodeKey]storage.Installation,
	actions map[NodeKey]string,
	opts ExplainOpts,
) (map[string]storage.Run, map[string]storage.JobStatus, error) {
	runs := make(map[string]storage.Run)
	statuses := make(map[string]storage.JobStatus, len(jobIDs))

	for key, jobID := range jobIDs {
		action := actions[key]
		if action == JobActionSkip {
			statuses[jobID] = storage.JobStatus{Status: cnab.StatusSucceeded}
			continue
		}

		node := g.Nodes[key]
		bun := node.Bundle
		if node.ResolvedInstallation != nil && action == cnab.ActionUpgrade {
			pulled, err := pullBundleForNode(ctx, p, key.Reference, opts)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot build the run for %s: %w", key, err)
			}
			bun = pulled
		}

		inst := installations[key]
		run := inst.NewRun(action, bun)
		run.Bundle = bun.Bundle
		run.BundleReference = bundleReferenceString(key, inst)

		runs[jobID] = run
		statuses[jobID] = storage.JobStatus{RunID: run.ID, Status: cnab.StatusPending}
	}

	return runs, statuses, nil
}
