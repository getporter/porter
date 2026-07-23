package cnab

import "sort"

// InterfaceRequirement is the set of capability names, by name only (v1 --
// no schema/type checking, no well-known-identifier matching, see #2650),
// that a candidate must provide to satisfy a bundle interface.
type InterfaceRequirement struct {
	Outputs     []string
	Parameters  []string
	Credentials []string
}

// InterfaceCandidate is the set of capability names a candidate actually
// offers. It may be built from a real bundle definition
// (NewInterfaceCandidateFromBundle) or, when no bundle definition is
// available (e.g. matching against an already-installed installation's
// recorded outputs), constructed directly from whatever names are known.
type InterfaceCandidate struct {
	Outputs     []string
	Parameters  []string
	Credentials []string
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
		Outputs:     mapKeys(b.Outputs),
		Parameters:  mapKeys(b.Parameters),
		Credentials: mapKeys(b.Credentials),
	}
}

func mapKeys[V any](m map[string]V) []string {
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

// missingNames returns the entries in required that are absent from
// available, or nil when none are missing.
func missingNames(required, available []string) []string {
	if len(required) == 0 {
		return nil
	}

	availableSet := make(map[string]bool, len(available))
	for _, name := range available {
		availableSet[name] = true
	}

	var missing []string
	for _, name := range required {
		if !availableSet[name] {
			missing = append(missing, name)
		}
	}
	return missing
}
