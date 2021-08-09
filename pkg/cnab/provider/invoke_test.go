package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
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
		run  claims.Run
		temp bool
		err  error
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
				Modifies:  true,
			},
		},
	}

	r := NewTestRuntime(t)
	defer r.Teardown()

	installation := r.TestClaims.CreateInstallation(claims.NewInstallation("dev", "exists"))
	lastRun := r.TestClaims.CreateRun(installation.NewRun(cnab.ActionInstall))

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
				run: claims.Run{
					Installation: "nonexistent",
					Bundle:       bun,
				},
			},
		},
		{
			name: "stateless action, existing claim should result in temp claim",
			in: input{
				bun,
				"blah",
				"exists",
			},
			want: result{
				run:  lastRun,
				temp: true,
			},
		},
		{
			name: "modifies and stateful action, existing claim should result in non temp claim",
			in: input{
				bun,
				"other",
				"exists",
			},
			want: result{
				run:  lastRun,
				temp: false,
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
				run:  lastRun,
				temp: false,
				err:  storage.ErrNotFound{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.in
			want := tc.want

			f, err := r.FileSystem.Create("bundle.json")
			require.NoError(t, err, "could not open bundle.json")
			_, err = tc.in.bun.WriteTo(f)
			require.NoError(t, err, "could not write to bundle.json")

			args := ActionArguments{
				Action:       in.action,
				Namespace:    "dev",
				Installation: in.installation,
				BundlePath:   "bundle.json",
			}
			runErr := r.Execute(args)
			c, claimErr := r.claims.GetLastRun(args.Namespace, args.Installation)

			if want.err == nil {
				require.NoErrorf(t, runErr, "Invoke failed")
				if want.temp {
					require.True(t, claimErr != nil || c.Action != args.Action, "temp claim should not be persisted")
				} else {
					require.NoError(t, claimErr, "the claim could not be read back")
					assert.Equal(t, in.action, c.Action, "claim saved with wrong action")
				}
			} else {
				require.ErrorIs(t, errors.Cause(runErr), want.err, "Invoke returned an unexpected error")
			}
		})
	}
}

func TestInvoke_NoClaimBubblesUpError(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Teardown()

	args := ActionArguments{
		Action:       "custom-action",
		Installation: "mybuns",
	}
	err := r.Execute(args)
	require.ErrorIs(t, err, storage.ErrNotFound{})
}
