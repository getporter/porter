package porter

import (
	"fmt"
	"sort"

	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
)

// wiringStrategy is the secrets.Source.Strategy used for a value sourced
// from another job's not-yet-produced output. Resolving it (looking up
// workflow.jobs.<jobID>.outputs.<name>, per
// v2.DependencySource.AsWorkflowWiring) is a #2647 concern -- this package
// only ever writes the reference, never resolves it.
const wiringStrategy = "porter"

// wireDependencyValues resolves dep's parameter or credential template map
// (dep.Parameters or dep.Credentials, passed as values) into a
// secrets.StrategyList: a literal value becomes a plain value-strategy
// entry (storage.ValueStrategy); a reference to the root bundle's own
// parameter or credential is resolved immediately from rootParameters/rootCredentials (kept separate
// since a parameter and credential may share a name), since
// the root's values are already known at workflow-generation time. A
// reference to a sibling dependency's output is skipped here -- it's
// handled by wireFromEdges instead, which reuses the graph's
// already-validated wiring edges rather than re-parsing the same
// reference.
func wireDependencyValues(values map[string]string, rootParameters, rootCredentials secrets.Set) (secrets.StrategyList, error) {
	if len(values) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	var wired secrets.StrategyList
	for _, name := range names {
		src, err := v2.ParseDependencySource(values[name])
		if err != nil {
			return nil, fmt.Errorf("cannot wire %q: %w", name, err)
		}

		switch {
		case src.Dependency != "":
			// Sibling-dependency-output reference: handled by wireFromEdges.
			continue
		case src.Parameter != "":
			value, ok := rootParameters[src.Parameter]
			if !ok {
				return nil, fmt.Errorf("cannot wire %q: root bundle parameter %q has no resolved value", name, src.Parameter)
			}
			wired = append(wired, storage.ValueStrategy(name, value))
		case src.Credential != "":
			value, ok := rootCredentials[src.Credential]
			if !ok {
				return nil, fmt.Errorf("cannot wire %q: root bundle credential %q has no resolved value", name, src.Credential)
			}
			wired = append(wired, storage.ValueStrategy(name, value))
		default:
			wired = append(wired, storage.ValueStrategy(name, src.Value))
		}
	}

	return wired, nil
}

// wireFromEdges returns the wiring entries for field ("parameters" or
// "credentials") sourced from a sibling dependency's output, using g's
// already-validated wiring edges (Graph.EdgesFrom(key), EdgeKindWiring)
// instead of re-parsing dep.Parameters/dep.Credentials -- GraphBuilder
// already extracted and deduped these (see extractWiringRefs in
// dependency_wiring.go).
func wireFromEdges(g *Graph, key NodeKey, jobIDs map[NodeKey]string, field string) secrets.StrategyList {
	var wired secrets.StrategyList
	for _, edge := range g.EdgesFrom(key) {
		if edge.Kind != EdgeKindWiring || edge.Detail == nil || edge.Detail.Field != field {
			continue
		}

		hint := v2.DependencySource{Dependency: edge.ToAlias, Output: edge.Detail.SourceOutput}.AsWorkflowWiring(jobIDs[edge.To])
		wired = append(wired, secrets.SourceMap{
			Name:   edge.Detail.FieldName,
			Source: secrets.Source{Strategy: wiringStrategy, Hint: hint},
		})
	}
	return wired
}

// wireJobParameters populates job.Installation.Parameters.Parameters from
// dep's Parameters template map, combining literal/root-parameter entries
// (wireDependencyValues) with sibling-output wiring entries
// (wireFromEdges).
func wireJobParameters(job *storage.Job, dep v2.Dependency, g *Graph, key NodeKey, jobIDs map[NodeKey]string, rootParameters, rootCredentials secrets.Set) error {
	wired, err := wireDependencyValues(dep.Parameters, rootParameters, rootCredentials)
	if err != nil {
		return err
	}
	job.Installation.Parameters.Parameters = append(job.Installation.Parameters.Parameters, wired...)
	job.Installation.Parameters.Parameters = append(job.Installation.Parameters.Parameters, wireFromEdges(g, key, jobIDs, "parameters")...)
	return nil
}

// wireJobCredentials is wireJobParameters' counterpart for dep.Credentials,
// populating job.Credentials instead of job.Installation.Parameters.
func wireJobCredentials(job *storage.Job, dep v2.Dependency, g *Graph, key NodeKey, jobIDs map[NodeKey]string, rootParameters, rootCredentials secrets.Set) error {
	wired, err := wireDependencyValues(dep.Credentials, rootParameters, rootCredentials)
	if err != nil {
		return err
	}
	job.Credentials = append(job.Credentials, wired...)
	job.Credentials = append(job.Credentials, wireFromEdges(g, key, jobIDs, "credentials")...)
	return nil
}

// propagateNamedSets copies parent's named credential/parameter sets onto
// a newly created dependency installation, so the sets already available
// to the parent are also available to satisfy any of the dependency's own
// requirements that aren't wired from elsewhere. Mirrors the existing
// convention in dependencies.go (runDependencyv2: "For now, assume it's
// okay to give the dependency the same credentials as the parent"). Only
// meant to be called for a brand-new dependency installation -- an
// existing one (Node.ResolvedInstallation) already has its own sets and
// shouldn't have them overwritten.
func propagateNamedSets(dep *storage.Installation, parent storage.Installation) {
	// Copy, don't share the parent's backing arrays.
	dep.CredentialSets = append([]string(nil), parent.CredentialSets...)
	dep.ParameterSets = append([]string(nil), parent.ParameterSets...)
}
