package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
)

// JobActionSkip marks a Job that performs no bundle action because the
// node it represents is already in sync with an existing, reusable
// installation. It exists so DependsOn/wiring references still have a
// stable job ID even when a dependency needs no work; WorkflowStatus.Jobs
// transitions such a job straight to a terminal status once a workflow run
// reaches it (a #2647 concern).
const JobActionSkip = "skip"

// ResolveNodeActions determines, for every node in g, which bundle action
// its Job should perform. rootAction is the action already chosen for the
// root bundle by the caller (e.g. porter install/upgrade) and is used
// as-is, not re-derived.
//
// A brand-new dependency node (no ResolvedInstallation) always gets
// cnab.ActionInstall. A node satisfied by an existing installation
// (Node.ResolvedInstallation, set by GraphBuilder via
// findExistingInstallation) is compared against its desired bundle
// reference: JobActionSkip if it already matches, cnab.ActionUpgrade
// otherwise. In practice findExistingInstallation only ever selects an
// installation whose digest already matches the desired reference (see
// candidateMatchesBundle in dependency_installation_resolver.go), so the
// upgrade branch isn't reachable through today's graph-building path --
// this comparison is kept as the correct, defensive check in case that
// matching criteria loosens in the future (e.g. matching on
// repository+sharing group alone).
//
// This only compares bundle digest, not resolved parameter/credential
// values -- a dependency whose only change is an upstream wired value is
// treated as in sync, since wiring (see workflow_wiring.go) is resolved
// only after actions are determined here. Known limitation, not a bug.
func ResolveNodeActions(g *Graph, rootAction string) (map[NodeKey]string, error) {
	actions := make(map[NodeKey]string, len(g.Nodes))

	for key, node := range g.Nodes {
		if key.IsRoot {
			actions[key] = rootAction
			continue
		}

		if node.ResolvedInstallation == nil {
			actions[key] = cnab.ActionInstall
			continue
		}

		ref, err := cnab.ParseOCIReference(key.Reference)
		if err != nil {
			return nil, fmt.Errorf("cannot determine the action for %s: %w", key, err)
		}

		if candidateMatchesBundle(*node.ResolvedInstallation, ref) {
			actions[key] = JobActionSkip
		} else {
			actions[key] = cnab.ActionUpgrade
		}
	}

	return actions, nil
}
