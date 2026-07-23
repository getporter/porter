package porter

import (
	"context"
	"regexp"

	"get.porter.sh/porter/pkg/cnab"
	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"get.porter.sh/porter/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
)

// selfOutputReference matches a dependency's Outputs map value that promotes
// one of its own outputs to its parent verbatim, e.g. "${outputs.connstr}".
// This is distinct from v2.ParseAllDependencySources, which only recognizes
// values prefixed with "bundle." (a sibling dependency's output, or the root
// bundle's own output) -- a bare "outputs.NAME" reference is this
// dependency's own output and never matched by that parser.
var selfOutputReference = regexp.MustCompile(`^\$\{\s*outputs\.([^\s}]+)\s*\}$`)

// requiredOutputNames returns the set of alias's own output names that must
// be present on a candidate installation for it to satisfy this dependency:
// outputs alias promotes to its parent via its own Outputs map, plus any
// output a sibling dependency wires from alias (refsByToAlias, keyed by
// wiringRef.ToAlias, as produced by extractWiringRefs).
func requiredOutputNames(alias string, dep v2.Dependency, refsByToAlias map[string][]wiringRef) []string {
	seen := make(map[string]bool)
	var names []string
	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		names = append(names, name)
	}

	for _, value := range dep.Outputs {
		if m := selfOutputReference.FindStringSubmatch(value); m != nil {
			add(m[1])
		}
	}

	for _, ref := range refsByToAlias[alias] {
		add(ref.Detail.SourceOutput)
	}

	return names
}

// findExistingInstallation looks for an already-installed installation, in
// namespace or the global namespace, that can be reused to satisfy lock
// instead of pulling and installing a new instance of its bundle. Returns
// nil (no error) when no candidate satisfies every rule.
//
// required may include names contributed by a declared bundle interface
// (see composeRequiredInterface), but only outputs are ever checked here
// regardless of what the interface declares -- this is the "existing
// installation" resolution context from PEP003 (see #2686), where
// credentials/parameters aren't re-evaluated because the installation is
// already running. Whether that should also hold for upgrade/uninstall/
// custom actions against an existing installation is an open question
// tracked by #2686, not resolved here.
func findExistingInstallation(ctx context.Context, p *Porter, namespace string, lock cnab.DependencyLock, required []string) (*storage.Installation, error) {
	if !lock.SharingMode {
		// A dependency with sharing disabled must never reuse an
		// installation, the same as it never collapses with another node
		// within a single graph.
		return nil, nil
	}

	ref, err := cnab.ParseOCIReference(lock.Reference)
	if err != nil {
		return nil, err
	}

	// The sharing group label's own name contains dots ("sh.porter.SharingGroup"),
	// which collide with the storage layer's dot-path notation for querying
	// into a nested field (it would be read as labels.sh.porter.SharingGroup,
	// a 3-level-deep path, not the flat key "sh.porter.SharingGroup") -- so
	// the label match happens in Go below instead of in the query filter.
	filter := bson.M{
		"uninstalled":       bson.M{"$ne": true},
		"bundle.repository": ref.Repository(),
		"$or": []bson.M{
			{"namespace": ""},
			{"namespace": namespace},
		},
	}
	// lock.Reference is normally digest-pinned (see candidateMatchesBundle),
	// so filtering on the recorded digest narrows candidates at the query
	// level in the common case instead of loading every installation of the
	// bundle to compare in Go.
	if ref.HasDigest() {
		filter["status.bundleDigest"] = ref.Digest().String()
	}

	query := storage.FindOptions{
		Sort:   []string{"-namespace"}, // prefer the local namespace match over the global ("") one
		Filter: filter,
	}

	candidates, err := p.Installations.FindInstallations(ctx, query)
	if err != nil {
		return nil, err
	}

	for i := range candidates {
		candidate := candidates[i]

		if candidate.Labels[sharingGroupLabel] != lock.SharingGroup {
			continue
		}

		if !candidate.IsInstalled() {
			continue
		}

		if !candidateMatchesBundle(candidate, ref) {
			continue
		}

		if !hasRequiredOutputs(ctx, p, candidate, required) {
			continue
		}

		return &candidate, nil
	}

	return nil, nil
}

// candidateMatchesBundle reports whether candidate is currently running the
// bundle identified by ref. A published bundle's dependency references are
// typically rewritten to a content digest at publish time (so lock.Reference
// here is usually digest-pinned, e.g. "repo@sha256:..."), while an
// installation is normally tracked by tag/version instead -- comparing the
// two reference strings directly would almost never match. So when ref
// carries a digest, this compares it against candidate's
// Status.BundleDigest, the actual digest of the bundle content last run,
// which is populated regardless of whether the installation was originally
// referenced by tag or digest. Otherwise (ref has no digest, e.g. an
// unpublished/local dependency reference) it falls back to comparing the
// full reference string against how candidate is tracked.
func candidateMatchesBundle(candidate storage.Installation, ref cnab.OCIReference) bool {
	candidateRef, ok, err := candidate.Bundle.GetBundleReference()
	if err != nil || !ok || candidateRef.Repository() != ref.Repository() {
		return false
	}

	if ref.HasDigest() {
		return candidate.Status.BundleDigest == ref.Digest().String()
	}

	return candidateRef.String() == ref.String()
}

// hasRequiredOutputs checks that every name in required is present among
// candidate's last recorded outputs.
func hasRequiredOutputs(ctx context.Context, p *Porter, candidate storage.Installation, required []string) bool {
	if len(required) == 0 {
		return true
	}

	outputs, err := p.Installations.GetLastOutputs(ctx, candidate.Namespace, candidate.Name)
	if err != nil {
		return false
	}

	available := make([]string, 0, outputs.Len())
	for _, o := range outputs.Value() {
		available = append(available, o.Name)
	}

	result := cnab.EvaluateInterfaceMatch(
		cnab.InterfaceCandidate{Outputs: available},
		cnab.InterfaceRequirement{Outputs: required},
		cnab.InterfaceMatchOutputsOnly,
	)
	return result.Satisfied
}
