package cnabprovider

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Runtime_GetClaimForInvoke(t *testing.T) {
	type input struct {
		bun    bundle.Bundle
		action string
		claim  string
	}

	type result struct {
		claim *claim.Claim
		temp  bool
		err   error
	}

	type test struct {
		name string
		in   input
		want result
	}

	bun := bundle.Bundle{
		Actions: map[string]bundle.Action{
			"blah": {
				Stateless: true,
			},
			"other": {
				Stateless: false,
			},
		},
	}

	eClaim, err := claim.New("exists", claim.ActionInstall, bun, nil)
	require.NoError(t, err)

	d := NewTestRuntime(t)

	err = d.claims.SaveClaim(eClaim)
	require.NoError(t, err)

	tests := []test{
		{
			name: "stateless action, no claim should result in temp claim",
			in: input{
				bun,
				"blah",
				"nonexistent",
			},
			want: result{
				temp: true,
				err:  nil,
				claim: &claim.Claim{
					Installation: "nonexistent",
					Bundle:       bun,
				},
			},
		},
		{
			name: "stateless action, existing claim should result in non temp claim",
			in: input{
				bun,
				"install",
				"exists",
			},
			want: result{
				claim: &eClaim,
				temp:  false,
				err:   nil,
			},
		},
		{
			name: "stateful action, existing claim should result in non temp claim",
			in: input{
				bun,
				"install",
				"exists",
			},
			want: result{
				claim: &eClaim,
				temp:  false,
				err:   nil,
			},
		},
		{
			name: "stateful action, non exist claim should result in error",
			in: input{
				bun,
				"other",
				"nonexist",
			},
			want: result{
				claim: &eClaim,
				temp:  false,
				err:   errors.Wrap(claim.ErrInstallationNotFound, "could not load installation nonexist"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.in
			want := tc.want
			c, temp, err := d.getClaimForInvoke(in.bun, in.action, in.claim)
			if want.err == nil {
				require.NoErrorf(t, err, "getClaimForInvoke failed")
				assert.Equalf(t, want.temp, temp, "getClaimForInvoke returned an unexpected temporary flag")
				assert.Equal(t, in.action, c.Action, "getClaimForInvoke returned a claim with an unexpected Action")
				if temp {
					assert.NotEqual(t, eClaim.ID, c.ID, "the temporary claim should have a new id")
				} else {
					assert.Equal(t, eClaim.ID, c.ID, "the claim should be a persisted claim")
				}
			} else {
				require.EqualErrorf(t, err, want.err.Error(), "getClaimForInvoke returned an unexpected error")
			}
		})
	}
}

func TestInvoke_NoClaimBubblesUpError(t *testing.T) {
	r := NewTestRuntime(t)

	args := ActionArguments{
		Installation: "mybuns",
	}
	err := r.Invoke("custom-action", args)
	require.EqualError(t, err, "could not load installation mybuns: Installation does not exist")
}
