package porter

import (
	"context"
	"errors"
	"maps"

	"get.porter.sh/porter/pkg/cnab"
	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"github.com/cnabio/cnab-go/bundle"
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
		Outputs: zeroOutputs(requiredOutputNames(alias, dep, refsByToAlias)),
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
		required.Outputs = mergeDefinitions(required.Outputs, candidate.Outputs)
		required.Parameters = mergeDefinitions(required.Parameters, candidate.Parameters)
		required.Credentials = mergeDefinitions(required.Credentials, candidate.Credentials)
		required.Definitions = mergeDefinitions(required.Definitions, candidate.Definitions)

	case hasDocument:
		// The document's own outputs/parameters/credentions already carry
		// real bundle.Output/Parameter/Credential definitions -- merge them
		// directly instead of reducing to names. Definitions is left as-is
		// (nil unless a reference also contributed one): a
		// DependencyInterfaceDocument has no Definitions map of its own, and
		// resolving its Definition name references would require the parent
		// bundle's own Definitions map, which isn't threaded through here
		// (see #2685 follow-up).
		required.Outputs = mergeDefinitions(required.Outputs, dep.Interface.Document.Outputs)
		required.Parameters = mergeDefinitions(required.Parameters, dep.Interface.Document.Parameters)
		required.Credentials = mergeDefinitions(required.Credentials, dep.Interface.Document.Credentials)
	}

	// ID-only (neither Reference nor Document set): nothing structural to
	// add. The ID itself isn't evaluated here -- see
	// cnab.EvaluateInterfaceMatch's doc comment for why, and #2686 for
	// where a whole-interface ID shortcut would apply.

	return required, nil
}

// zeroOutputs wraps names (the base interface, inferred from wiring/usage)
// as zero-value bundle.Output entries -- there is no schema to attach on the
// base side, only the fact that an output with this name must exist. Returns
// nil, not an empty map, when names is empty, so composing with no declared
// interface still produces the exact same value as before this type widened.
func zeroOutputs(names []string) map[string]bundle.Output {
	if len(names) == 0 {
		return nil
	}

	m := make(map[string]bundle.Output, len(names))
	for _, name := range names {
		m[name] = bundle.Output{}
	}
	return m
}

// mergeDefinitions merges overlay on top of base: overlay's definitions win
// on a name collision, per PEP003's "declared interface merged on top of
// the base interface" (see #2685). The returned maps -- and the values
// within them -- are borrowed from base/overlay, not deep-copied.
func mergeDefinitions[T any](base, overlay map[string]T) map[string]T {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}

	merged := make(map[string]T, len(base)+len(overlay))
	maps.Copy(merged, base)
	maps.Copy(merged, overlay)
	return merged
}
