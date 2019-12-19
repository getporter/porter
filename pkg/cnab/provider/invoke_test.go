package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	instancestorage "get.porter.sh/porter/pkg/instance-storage"
	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/cnab-go/claim"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ClaimWriting(t *testing.T) {

	type op struct {
		bun    *bundle.Bundle
		action string
		claim  string
	}
	type test struct {
		name   string
		in     op
		status string
		want   bool
	}

	c := config.NewTestConfig(t)
	instanceStorage := instancestorage.NewTestInstanceStorageProvider()
	d := NewRuntime(c.Config, instanceStorage)

	eClaim, err := claim.New("exists")
	require.NoError(t, err)
	eClaim.Update(claim.ActionInstall, claim.StatusSuccess)

	err = instanceStorage.Store(*eClaim)
	require.NoError(t, err)

	bun := &bundle.Bundle{
		Actions: map[string]bundle.Action{
			"blah": bundle.Action{
				Stateless: true,
			},
			"other": bundle.Action{
				Stateless: false,
			},
		},
	}

	tests := []test{
		{
			name: "stateless action, no claim should result in temp claim not written",
			in: op{
				bun,
				"blah",
				"nonexistent",
			},
			status: claim.StatusFailure,
			want:   false,
		},
		{
			name: "stateless action, existing claim should result in non temp claim and should be written",
			in: op{
				bun,
				"blah",
				"exists",
			},
			status: claim.StatusFailure,
			want:   true,
		},
		{
			name: "stateful action, existing claim should result in non temp claim and should be written",
			in: op{
				bun,
				"other",
				"exists",
			},
			status: claim.StatusFailure,
			want:   true,
		},
	}

	for _, tc := range tests {
		in := tc.in
		c, temp, err := d.getClaim(in.bun, in.action, in.claim)
		require.NoError(t, err)
		c.Result.Action = in.action
		c.Result.Status = tc.status
		err = d.writeClaim(temp, c)
		assert.NoError(t, err)

		fc, err := d.instanceStorage.Read(in.claim)
		if tc.want {
			assert.NoErrorf(t, err, "expected claim for %s", tc.name)
			assert.Equalf(t, in.action, fc.Result.Action, "expected action=%s for %s", in.action, tc.name)
			assert.Equalf(t, tc.status, fc.Result.Status, "expected status=%s for %s", tc.status, tc.name)
		} else {
			assert.Error(t, err, "expected no claim for %s", tc.name)
		}
	}
}

func Test_ClaimLoading(t *testing.T) {
	type input struct {
		bun    *bundle.Bundle
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

	bun := &bundle.Bundle{
		Actions: map[string]bundle.Action{
			"blah": bundle.Action{
				Stateless: true,
			},
			"other": bundle.Action{
				Stateless: false,
			},
		},
	}

	eClaim, err := claim.New("exists")
	require.NoError(t, err)
	eClaim.Update(claim.ActionInstall, claim.StatusSuccess)

	c := config.NewTestConfig(t)
	instanceStorage := instancestorage.NewTestInstanceStorageProvider()
	d := NewRuntime(c.Config, instanceStorage)

	err = instanceStorage.Store(*eClaim)
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
					Name:   "nonexistent",
					Bundle: bun,
				},
			},
		},
		{
			name: "stateless action, existing claim should result in non temp claim",
			in: input{
				bun,
				"blah",
				"exists",
			},
			want: result{
				claim: eClaim,
				temp:  false,
				err:   nil,
			},
		},
		{
			name: "stateful action, existing claim should result in non temp claim",
			in: input{
				bun,
				"other",
				"exists",
			},
			want: result{
				claim: eClaim,
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
				claim: eClaim,
				temp:  false,
				err:   errors.Wrap(claim.ErrClaimNotFound, "could not load claim nonexist"),
			},
		},
	}

	for _, tc := range tests {
		in := tc.in
		want := tc.want
		_, temp, err := d.getClaim(in.bun, in.action, in.claim)
		assert.Equalf(t, want.temp, temp, "%s: expected temp=want.temp", tc.name)
		if want.err == nil {
			assert.NoErrorf(t, err, "%s: expected no error", tc.name)
		} else {
			assert.Errorf(t, err, "%s: expected error", tc.name)
			assert.EqualErrorf(t, want.err, err.Error(), "%s: expected error %s, got %s", tc.name, want.err, err)
		}
	}
}

func Test_Invoke_Params(t *testing.T) {
	testcases := []struct {
		name            string
		required        bool
		provided        bool
		defaultExists   bool
		appliesToAction bool
		expectedVal     interface{}
		expectedErr     string
	}{
		{"required, provided, default exists, applies to action",
			true, true, true, true, "my-param-value", "",
		},
		{"required, provided, default exists, does not apply to action",
			true, true, true, false, "my-param-value", "",
		},
		{"required, provided, default does not exist, applies to action",
			true, true, false, true, "my-param-value", "",
		},
		{"required, provided, default does not exist, does not apply to action",
			true, true, false, false, "my-param-value", "",
		},
		{"required, not provided, default exists, applies to action",
			true, false, true, true, "my-param-default", "",
		},
		{"required, not provided, default exists, does not apply to action",
			true, false, true, false, "my-param-default", "",
		},
		{"required, not provided, default does not exist, applies to action",
			true, false, false, true, nil, "invalid parameters: parameter \"my-param\" is required",
		},
		{"required, not provided, default does not exist, does not apply to action",
			true, false, false, false, "", "",
		},
		{"not required, provided, default exists, applies to action",
			false, true, true, true, "my-param-value", "",
		},
		{"not required, provided, default exists, does not apply to action",
			false, true, true, false, "my-param-value", "",
		},
		{"not required, provided, default does not exist, applies to action",
			false, true, false, true, "my-param-value", "",
		},
		{"not required, provided, default does not exist, does not apply to action",
			false, true, false, false, "my-param-value", "",
		},
		{"not required, not provided, default exists, applies to action",
			false, false, true, true, "my-param-default", "",
		},
		{"not required, not provided, default exists, does not apply to action",
			false, false, true, false, "my-param-default", "",
		},
		{"not required, not provided, default does not exist, applies to action",
			false, false, false, true, nil, "",
		},
		{"not required, not provided, default does not exist, does not apply to action",
			false, false, false, false, nil, "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := config.NewTestConfig(t)
			instanceStorage := instancestorage.NewTestInstanceStorageProvider()
			d := NewRuntime(c.Config, instanceStorage)

			bun := &bundle.Bundle{
				InvocationImages: []bundle.InvocationImage{
					{
						BaseImage: bundle.BaseImage{
							Image:     "mybuns:latest",
							ImageType: "docker",
						},
					},
				},
				Actions: map[string]bundle.Action{
					"zombies": bundle.Action{
						Modifies: true,
					},
				},
				Definitions: definition.Definitions{
					"my-param": &definition.Schema{
						Type: "string",
					},
				},
				Parameters: map[string]bundle.Parameter{
					"my-param": bundle.Parameter{
						Definition: "my-param",
						Required:   tc.required,
					},
				},
			}

			if tc.defaultExists {
				bun.Definitions["my-param"].Default = "my-param-default"
			}

			if !tc.appliesToAction {
				param := bun.Parameters["my-param"]
				param.ApplyTo = []string{"not-invoke"}
				bun.Parameters["my-param"] = param
			}

			claim, err := claim.New("test")
			require.NoError(t, err)

			claim.Bundle = bun
			d.instanceStorage.Store(*claim)

			args := ActionArguments{
				Claim:    "test",
				Insecure: true,
				Driver:   "debug",
			}
			if tc.provided {
				args.Params = map[string]string{
					"my-param": "my-param-value",
				}
			}

			err = d.Invoke("zombies", args)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}

			updatedClaim, err := d.instanceStorage.Read("test")
			require.NoError(t, err)
			require.Equal(t, tc.expectedVal, updatedClaim.Parameters["my-param"])
		})
	}
}
