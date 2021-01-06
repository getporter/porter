package cnabprovider

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_ClaimPersistence(t *testing.T) {
	t.Parallel()

	type input struct {
		bun          bundle.Bundle
		action       string
		installation string
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

	bunV, _ := bundle.GetDefaultSchemaVersion()
	bun := bundle.Bundle{
		SchemaVersion:    bunV,
		InvocationImages: []bundle.InvocationImage{{BaseImage: bundle.BaseImage{Image: "example.com/foo:v1.0.0"}}},
		Actions: map[string]bundle.Action{
			"blah": {
				Stateless: true,
			},
			"other": {
				Stateless: false,
			},
		},
	}

	d := NewTestRuntime(t)
	eClaim := d.TestClaims.CreateClaim("exists", claim.ActionInstall, bun, nil)

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

			f, err := d.FileSystem.Create("bundle.json")
			require.NoError(t, err, "could not open bundle.json")
			_, err = tc.in.bun.WriteTo(f)
			require.NoError(t, err, "could not write to bundle.json")

			args := ActionArguments{
				Action:       in.action,
				Installation: in.installation,
				BundlePath:   "bundle.json",
			}
			runErr := d.Execute(args)
			c, claimErr := d.claims.ReadLastClaim(in.installation)

			if want.err == nil {
				require.NoErrorf(t, runErr, "Invoke failed")
				if want.temp {
					require.Error(t, claimErr, "temp claim should not be persisted")
				} else {
					require.NoError(t, claimErr, "the claim could not be read back")
					assert.Equal(t, in.action, c.Action, "claim saved with wrong action")
				}
			} else {
				require.EqualErrorf(t, runErr, want.err.Error(), "Invoke returned an unexpected error")
			}
		})
	}
}

func TestInvoke_NoClaimBubblesUpError(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)

	args := ActionArguments{
		Action:       "custom-action",
		Installation: "mybuns",
	}
	err := r.Execute(args)
	require.EqualError(t, err, "could not load installation mybuns: Installation does not exist")
}
