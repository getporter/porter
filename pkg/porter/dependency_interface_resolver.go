package porter

import (
	"context"
	"errors"
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
)

// errInterfaceReferenceAndDocument is returned by composeRequiredInterface
// when a dependency declares both interface.reference and
// interface.document -- manifest.BundleInterface's own doc comment already
// states these are mutually exclusive, but nothing has enforced it up to
// now (see #2626). Unlike a transient interface.Reference pull failure,
// this is an authoring bug and must surface as a real failure rather than
// being treated as a best-effort "no match, fall back to pull."
var errInterfaceReferenceAndDocument = errors.New("dependency interface declares both a reference and a document; only one may be set")

// composeRequiredInterface computes the full InterfaceRequirement for a v2
// dependency: the union of outputs the parent bundle actually uses from it
// (requiredOutputNames) and whatever interface.Reference or
// interface.Document adds on top, per PEP003's bundle interface
// composition (https://github.com/getporter/proposals/blob/main/pep/003-advanced-dependencies.md#bundle-interfaces).
//
// Callers must check experimental.FlagDependenciesV2 before invoking this
// for a dependency with a non-nil Interface -- this function does not
// re-check it, matching the codebase's existing convention of gating
// experimental behavior at the call site (see dependency_graph_builder.go).
func (b *GraphBuilder) composeRequiredInterface(ctx context.Context, alias string, dep v2.Dependency, refsByToAlias map[string][]wiringRef, opts ExplainOpts) (cnab.InterfaceRequirement, error) {
	required := cnab.InterfaceRequirement{
		Outputs: requiredOutputNames(alias, dep, refsByToAlias),
	}

	if dep.Interface == nil {
		return required, nil
	}

	hasReference := dep.Interface.Reference != ""
	hasDocument := !dep.Interface.Document.IsEmpty()

	switch {
	case hasReference && hasDocument:
		return cnab.InterfaceRequirement{}, errInterfaceReferenceAndDocument

	case hasReference:
		referenceBun, err := b.pullDependencyBundle(ctx, dep.Interface.Reference, opts)
		if err != nil {
			return cnab.InterfaceRequirement{}, err
		}
		candidate := cnab.NewInterfaceCandidateFromBundle(referenceBun)
		required.Outputs = unionSorted(required.Outputs, candidate.Outputs)
		required.Parameters = unionSorted(required.Parameters, candidate.Parameters)
		required.Credentials = unionSorted(required.Credentials, candidate.Credentials)

	case hasDocument:
		outputs, parameters, credentials := dep.Interface.Document.Names()
		required.Outputs = unionSorted(required.Outputs, outputs)
		required.Parameters = unionSorted(required.Parameters, parameters)
		required.Credentials = unionSorted(required.Credentials, credentials)
	}

	// ID-only (neither Reference nor Document set): nothing structural to
	// add. The ID itself isn't evaluated here -- see
	// cnab.EvaluateInterfaceMatch's doc comment for why, and #2686 for
	// where a whole-interface ID shortcut would apply.

	return required, nil
}

// unionSorted returns the sorted, deduplicated union of a and b.
func unionSorted(a, b []string) []string {
	if len(b) == 0 {
		return a
	}

	seen := make(map[string]bool, len(a)+len(b))
	merged := make([]string, 0, len(a)+len(b))
	for _, name := range a {
		if !seen[name] {
			seen[name] = true
			merged = append(merged, name)
		}
	}
	for _, name := range b {
		if !seen[name] {
			seen[name] = true
			merged = append(merged, name)
		}
	}

	sort.Strings(merged)
	return merged
}
