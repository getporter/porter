package cnab

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
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
			candidate: InterfaceCandidate{Outputs: []string{"connstr"}},
			required:  InterfaceRequirement{Outputs: []string{"connstr"}},
			mode:      InterfaceMatchOutputsOnly,
			want:      InterfaceMatchResult{Satisfied: true},
		},
		{
			name:      "outputs only: candidate is missing the required output",
			candidate: InterfaceCandidate{Outputs: []string{"port"}},
			required:  InterfaceRequirement{Outputs: []string{"connstr"}},
			mode:      InterfaceMatchOutputsOnly,
			want: InterfaceMatchResult{
				Satisfied:      false,
				MissingOutputs: []string{"connstr"},
			},
		},
		{
			name: "outputs only: missing parameters and credentials are ignored",
			candidate: InterfaceCandidate{
				Outputs: []string{"connstr"},
			},
			required: InterfaceRequirement{
				Outputs:     []string{"connstr"},
				Parameters:  []string{"logLevel"},
				Credentials: []string{"token"},
			},
			mode: InterfaceMatchOutputsOnly,
			want: InterfaceMatchResult{Satisfied: true},
		},
		{
			name: "full mode: all present",
			candidate: InterfaceCandidate{
				Outputs:     []string{"connstr"},
				Parameters:  []string{"logLevel"},
				Credentials: []string{"token"},
			},
			required: InterfaceRequirement{
				Outputs:     []string{"connstr"},
				Parameters:  []string{"logLevel"},
				Credentials: []string{"token"},
			},
			mode: InterfaceMatchFull,
			want: InterfaceMatchResult{Satisfied: true},
		},
		{
			name: "full mode: missing parameter fails the match",
			candidate: InterfaceCandidate{
				Outputs:     []string{"connstr"},
				Credentials: []string{"token"},
			},
			required: InterfaceRequirement{
				Outputs:     []string{"connstr"},
				Parameters:  []string{"logLevel"},
				Credentials: []string{"token"},
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
				Outputs:    []string{"connstr"},
				Parameters: []string{"logLevel"},
			},
			required: InterfaceRequirement{
				Outputs:     []string{"connstr"},
				Parameters:  []string{"logLevel"},
				Credentials: []string{"token"},
			},
			mode: InterfaceMatchFull,
			want: InterfaceMatchResult{
				Satisfied:          false,
				MissingCredentials: []string{"token"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := EvaluateInterfaceMatch(tc.candidate, tc.required, tc.mode)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestNewInterfaceCandidateFromBundle(t *testing.T) {
	t.Parallel()

	bun := ExtendedBundle{Bundle: bundle.Bundle{
		Outputs: map[string]bundle.Output{
			"port": {}, "connstr": {},
		},
		Parameters: map[string]bundle.Parameter{
			"logLevel": {},
		},
		Credentials: map[string]bundle.Credential{
			"token": {}, "apiKey": {},
		},
	}}

	got := NewInterfaceCandidateFromBundle(bun)
	require.Equal(t, InterfaceCandidate{
		Outputs:     []string{"connstr", "port"},
		Parameters:  []string{"logLevel"},
		Credentials: []string{"apiKey", "token"},
	}, got)
}
