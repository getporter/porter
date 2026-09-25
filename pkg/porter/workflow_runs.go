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
func pullBundleForNode(ctx context.Context, p *Porter, ref string, opts ExplainOpts) (cnab.BundleReference, error) {
	pullOpts := BundlePullOptions{
		Reference:        ref,
		InsecureRegistry: opts.InsecureRegistry,
		Force:            opts.Force,
	}

	cachedBundle, err := p.PullBundle(ctx, pullOpts)
	if err != nil {
		return cnab.BundleReference{}, fmt.Errorf("failed to pull bundle %s: %w", ref, err)
	}

	return cachedBundle.BundleReference, nil
}

// buildJobRuns creates a Pending storage.Run for every job whose action
// requires one (i.e. not JobActionSkip), and a WorkflowStatus.JobStatus
// entry for every job in the graph, including skipped ones (marked
// succeeded immediately, since they require no bundle action).
//
// installations and actions must be the results of buildJobInstallations
// and ResolveNodeActions for the same graph/jobIDs. rootRef is the
// resolved reference (and digest) of the root bundle, as the normal
// lifecycle path records; the root Node carries only the definition.
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
	rootRef cnab.BundleReference,
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
		inst := installations[key]
		bun := node.Bundle
		bundleReference := key.Reference
		var bundleDigest string
		if key.IsRoot {
			// The root's resolved reference/digest, not the installation's
			// tracked ones, which can be stale until the run's result is
			// applied (e.g. an upgrade by version).
			bundleReference = rootRef.Reference.String()
			bundleDigest = rootRef.Digest.String()
		} else if node.ResolvedInstallation != nil && action == cnab.ActionUpgrade {
			pulled, err := pullBundleForNode(ctx, p, key.Reference, opts)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot build the run for %s: %w", key, err)
			}
			bun = pulled.Definition
			bundleDigest = pulled.Digest.String()
		} else {
			// Digest resolved by the pull that produced node.Bundle, so the two
			// always describe the same bundle version (a moving tag isn't
			// re-resolved here).
			bundleDigest = node.Digest
			if bundleDigest == "" {
				if ref, err := cnab.ParseOCIReference(key.Reference); err == nil && ref.HasDigest() {
					bundleDigest = ref.Digest().String()
				}
			}
		}

		run := inst.NewRun(action, bun)
		run.Bundle = bun.Bundle
		run.BundleReference = bundleReference
		run.BundleDigest = bundleDigest

		runs[jobID] = run
		statuses[jobID] = storage.JobStatus{RunID: run.ID, Status: cnab.StatusPending}
	}

	return runs, statuses, nil
}
