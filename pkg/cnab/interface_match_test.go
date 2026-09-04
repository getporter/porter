package cnab

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/require"
)

func TestEvaluateInterfaceMatch(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name      string
		candidate InterfaceCandidate
		required  InterfaceRequirement
		mode      InterfaceMatchMode
		want      InterfaceMatchResult
	}{
		{
			name:     "empty requirement is always satisfied",
			required: InterfaceRequirement{},
			mode:     InterfaceMatchFull,
			want:     InterfaceMatchResult{Satisfied: true},
		},
		{
			name:      "outputs only: candidate has the required output",
			candidate: InterfaceCandidate{Outputs: map[string]bundle.Output{"connstr": {}}},
			required:  InterfaceRequirement{Outputs: map[string]bundle.Output{"connstr": {}}},
			mode:      InterfaceMatchOutputsOnly,
			want:      InterfaceMatchResult{Satisfied: true},
		},
		{
			name:      "outputs only: candidate is missing the required output",
			candidate: InterfaceCandidate{Outputs: map[string]bundle.Output{"port": {}}},
			required:  InterfaceRequirement{Outputs: map[string]bundle.Output{"connstr": {}}},
			mode:      InterfaceMatchOutputsOnly,
			want: InterfaceMatchResult{
				Satisfied:      false,
				MissingOutputs: []string{"connstr"},
			},
		},
		{
			name: "outputs only: missing parameters and credentials are ignored",
			candidate: InterfaceCandidate{
				Outputs: map[string]bundle.Output{"connstr": {}},
			},
			required: InterfaceRequirement{
				Outputs:     map[string]bundle.Output{"connstr": {}},
				Parameters:  map[string]bundle.Parameter{"logLevel": {}},
				Credentials: map[string]bundle.Credential{"token": {}},
			},
			mode: InterfaceMatchOutputsOnly,
			want: InterfaceMatchResult{Satisfied: true},
		},
		{
			name: "full mode: all present",
			candidate: InterfaceCandidate{
				Outputs:     map[string]bundle.Output{"connstr": {}},
				Parameters:  map[string]bundle.Parameter{"logLevel": {}},
				Credentials: map[string]bundle.Credential{"token": {}},
			},
			required: InterfaceRequirement{
				Outputs:     map[string]bundle.Output{"connstr": {}},
				Parameters:  map[string]bundle.Parameter{"logLevel": {}},
				Credentials: map[string]bundle.Credential{"token": {}},
			},
			mode: InterfaceMatchFull,
			want: InterfaceMatchResult{Satisfied: true},
		},
		{
			name: "full mode: missing parameter fails the match",
			candidate: InterfaceCandidate{
				Outputs:     map[string]bundle.Output{"connstr": {}},
				Credentials: map[string]bundle.Credential{"token": {}},
			},
			required: InterfaceRequirement{
				Outputs:     map[string]bundle.Output{"connstr": {}},
				Parameters:  map[string]bundle.Parameter{"logLevel": {}},
				Credentials: map[string]bundle.Credential{"token": {}},
			},
			mode: InterfaceMatchFull,
			want: InterfaceMatchResult{
				Satisfied:         false,
				MissingParameters: []string{"logLevel"},
			},
		},
		{
			name: "full mode: missing credential fails the match",
			candidate: InterfaceCandidate{
				Outputs:    map[string]bundle.Output{"connstr": {}},
				Parameters: map[string]bundle.Parameter{"logLevel": {}},
			},
			required: InterfaceRequirement{
				Outputs:     map[string]bundle.Output{"connstr": {}},
				Parameters:  map[string]bundle.Parameter{"logLevel": {}},
				Credentials: map[string]bundle.Credential{"token": {}},
			},
			mode: InterfaceMatchFull,
			want: InterfaceMatchResult{
				Satisfied:          false,
				MissingCredentials: []string{"token"},
			},
		},
		{
			name: "full mode: shared name with differing definitions still matches -- matching is key-presence only, not value comparison",
			candidate: InterfaceCandidate{
				Outputs:     map[string]bundle.Output{"connstr": {Definition: "candidateDef", Path: "/cnab/app/outputs/connstr"}},
				Parameters:  map[string]bundle.Parameter{"logLevel": {Definition: "candidateDef", Required: false}},
				Credentials: map[string]bundle.Credential{"token": {Description: "candidate token"}},
			},
			required: InterfaceRequirement{
				Outputs:     map[string]bundle.Output{"connstr": {Definition: "requiredDef"}},
				Parameters:  map[string]bundle.Parameter{"logLevel": {Definition: "requiredDef", Required: true}},
				Credentials: map[string]bundle.Credential{"token": {Description: "required token"}},
			},
			mode: InterfaceMatchFull,
			want: InterfaceMatchResult{Satisfied: true},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := EvaluateInterfaceMatch(test.candidate, test.required, test.mode)
			require.Equal(t, test.want, got)
		})
	}
}

func TestNewInterfaceCandidateFromBundle(t *testing.T) {
	t.Parallel()

	defs := definition.Definitions{
		"connstrDef": &definition.Schema{Type: "string"},
	}
	bun := ExtendedBundle{Bundle: bundle.Bundle{
		Outputs: map[string]bundle.Output{
			"port":    {},
			"connstr": {Definition: "connstrDef", Path: "/cnab/app/outputs/connstr"},
		},
		Parameters: map[string]bundle.Parameter{
			"logLevel": {Required: true},
		},
		Credentials: map[string]bundle.Credential{
			"token": {}, "apiKey": {},
		},
		Definitions: defs,
	}}

	got := NewInterfaceCandidateFromBundle(bun)
	require.Equal(t, InterfaceCandidate{
		Outputs: map[string]bundle.Output{
			"port":    {},
			"connstr": {Definition: "connstrDef", Path: "/cnab/app/outputs/connstr"},
		},
		Parameters: map[string]bundle.Parameter{
			"logLevel": {Required: true},
		},
		Credentials: map[string]bundle.Credential{
			"token": {}, "apiKey": {},
		},
		Definitions: defs,
	}, got)
}

func TestInterfaceCandidate_OutputsHash(t *testing.T) {
	t.Parallel()

	t.Run("empty outputs hashes to an empty string", func(t *testing.T) {
		t.Parallel()

		require.Empty(t, InterfaceCandidate{}.OutputsHash())
	})

	t.Run("stable regardless of map iteration order", func(t *testing.T) {
		t.Parallel()

		a := InterfaceCandidate{Outputs: map[string]bundle.Output{"connstr": {}, "port": {}}}
		b := InterfaceCandidate{Outputs: map[string]bundle.Output{"port": {}, "connstr": {}}}

		require.NotEmpty(t, a.OutputsHash())
		require.Equal(t, a.OutputsHash(), b.OutputsHash())
	})

	t.Run("changes when the output name set changes", func(t *testing.T) {
		t.Parallel()

		a := InterfaceCandidate{Outputs: map[string]bundle.Output{"connstr": {}}}
		b := InterfaceCandidate{Outputs: map[string]bundle.Output{"connstr": {}, "port": {}}}

		require.NotEqual(t, a.OutputsHash(), b.OutputsHash())
	})

	t.Run("unaffected by output definitions -- name-only, matching EvaluateInterfaceMatch", func(t *testing.T) {
		t.Parallel()

		a := InterfaceCandidate{Outputs: map[string]bundle.Output{"connstr": {Definition: "defA", Path: "/a"}}}
		b := InterfaceCandidate{Outputs: map[string]bundle.Output{"connstr": {Definition: "defB", Path: "/b"}}}

		require.Equal(t, a.OutputsHash(), b.OutputsHash())
	})
}
