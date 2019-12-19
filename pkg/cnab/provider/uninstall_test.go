package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	instancestorage "get.porter.sh/porter/pkg/instance-storage"
	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func Test_Uninstall_Params(t *testing.T) {
	testcases := []struct {
		name            string
		required        bool
		provided        bool
		defaultExists   bool
		appliesToAction bool
		expectedErr     string
	}{
		{"required, provided, default exists, applies to action",
			true, true, true, true, "",
		},
		{"required, provided, default exists, does not apply to action",
			true, true, true, false, "",
		},
		{"required, provided, default does not exist, applies to action",
			true, true, false, true, "",
		},
		{"required, provided, default does not exist, does not apply to action",
			true, true, false, false, "",
		},
		{"required, not provided, default exists, applies to action",
			true, false, true, true, "",
		},
		{"required, not provided, default exists, does not apply to action",
			true, false, true, false, "",
		},
		{"required, not provided, default does not exist, applies to action",
			true, false, false, true, "invalid parameters: parameter \"my-param\" is required",
		},
		{"required, not provided, default does not exist, does not apply to action",
			true, false, false, false, "",
		},
		{"not required, provided, default exists, applies to action",
			false, true, true, true, "",
		},
		{"not required, provided, default exists, does not apply to action",
			false, true, true, false, "",
		},
		{"not required, provided, default does not exist, applies to action",
			false, true, false, true, "",
		},
		{"not required, provided, default does not exist, does not apply to action",
			false, true, false, false, "",
		},
		{"not required, not provided, default exists, applies to action",
			false, false, true, true, "",
		},
		{"not required, not provided, default exists, does not apply to action",
			false, false, true, false, "",
		},
		{"not required, not provided, default does not exist, applies to action",
			false, false, false, true, "",
		},
		{"not required, not provided, default does not exist, does not apply to action",
			false, false, false, false, "",
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
				param.ApplyTo = []string{"not-uninstall"}
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

			err = d.Uninstall(args)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
