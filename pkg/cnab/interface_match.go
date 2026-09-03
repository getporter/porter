package cnab

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
)

// InterfaceRequirement is the set of capabilities, keyed by name, that a
// candidate must provide to satisfy a bundle interface. Matching itself is
// still name-only (v1 -- no schema/type checking, no well-known-identifier
// matching, see #2650); the definitions are carried through so that future
// work has real data to check against instead of bare names.
//
// Definitions is only ever populated from an interface.reference (the
// pulled bundle's own Definitions map) -- interface.document has no
// Definitions of its own to carry, since its schema references, if any, are
// only resolvable against the parent bundle's Definitions map, which isn't
// threaded through here (see #2685 follow-up).
type InterfaceRequirement struct {
	Outputs     map[string]bundle.Output
	Parameters  map[string]bundle.Parameter
	Credentials map[string]bundle.Credential
	Definitions definition.Definitions
}

// InterfaceCandidate is the set of capabilities a candidate actually offers.
// It may be built from a real bundle definition
// (NewInterfaceCandidateFromBundle) or, when no bundle definition is
// available (e.g. matching against an already-installed installation's
// recorded outputs), constructed directly from whatever names are known.
//
// The maps returned by NewInterfaceCandidateFromBundle are borrowed from the
// source bundle, not copied -- callers must not mutate them.
type InterfaceCandidate struct {
	Outputs     map[string]bundle.Output
	Parameters  map[string]bundle.Parameter
	Credentials map[string]bundle.Credential
	Definitions definition.Definitions
}

// InterfaceMatchMode controls which capability categories must match,
// corresponding to the resolution contexts described in PEP003 (see #2686):
type InterfaceMatchMode int

const (
	// InterfaceMatchOutputsOnly requires only outputs to match. Used when
	// resolving against an arbitrary/user-supplied bundle (unmatched
	// credential/parameter mappings are ignored) or an already-installed
	// installation (only its outputs are known/relevant).
	InterfaceMatchOutputsOnly InterfaceMatchMode = iota

	// InterfaceMatchFull requires outputs, parameters, and credentials to
	// all match. Used when resolving the dependency's default declared
	// bundle/version, where the dependency's own credential/parameter
	// mappings are meant specifically for that bundle.
	InterfaceMatchFull
)

// InterfaceMatchResult is the outcome of evaluating a candidate against a
// requirement. Satisfied is the single bool callers branch on; the Missing*
// fields exist purely for diagnostics (error/warning messages), not as a
// score -- matching is binary, matching PEP003's deterministic resolution
// precedence (see #2626).
type InterfaceMatchResult struct {
	Satisfied bool

	MissingOutputs     []string
	MissingParameters  []string
	MissingCredentials []string
}

// NewInterfaceCandidateFromBundle builds an InterfaceCandidate from a real
// bundle definition's own outputs, parameters, and credentials.
func NewInterfaceCandidateFromBundle(b ExtendedBundle) InterfaceCandidate {
	return InterfaceCandidate{
		Outputs:     b.Outputs,
		Parameters:  b.Parameters,
		Credentials: b.Credentials,
		Definitions: b.Definitions,
	}
}

// OutputsHash returns a stable sha256 digest of the candidate's output
// names (sorted, name-only -- matching how EvaluateInterfaceMatch itself
// compares outputs, not their types/schemas). Empty when there are no
// outputs, matching the ParametersDigest/CredentialsDigest convention in
// storage.Run. Suitable for persisting on Installation.Status to cheaply
// compare installation interfaces later.
func (c InterfaceCandidate) OutputsHash() string {
	if len(c.Outputs) == 0 {
		return ""
	}
	data, _ := json.Marshal(SortedKeys(c.Outputs))
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data))
}

// SortedKeys returns the sorted keys of m.
func SortedKeys[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// EvaluateInterfaceMatch reports whether candidate satisfies required, per
// mode. It is a binary predicate (satisfies / doesn't) -- deliberately not
// a score, matching PEP003's fully deterministic resolution precedence.
func EvaluateInterfaceMatch(candidate InterfaceCandidate, required InterfaceRequirement, mode InterfaceMatchMode) InterfaceMatchResult {
	result := InterfaceMatchResult{
		MissingOutputs: missingNames(required.Outputs, candidate.Outputs),
	}

	if mode == InterfaceMatchFull {
		result.MissingParameters = missingNames(required.Parameters, candidate.Parameters)
		result.MissingCredentials = missingNames(required.Credentials, candidate.Credentials)
	}

	result.Satisfied = len(result.MissingOutputs) == 0 &&
		len(result.MissingParameters) == 0 &&
		len(result.MissingCredentials) == 0

	return result
}

// missingNames returns the keys of required that are absent from available,
// or nil when none are missing. Only key presence is compared -- values are
// never inspected, so matching stays name-only regardless of what richer
// data a definition carries (see #2650 for value/schema comparison).
func missingNames[T any](required map[string]T, available map[string]T) []string {
	if len(required) == 0 {
		return nil
	}

	var missing []string
	for name := range required {
		if _, ok := available[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}
