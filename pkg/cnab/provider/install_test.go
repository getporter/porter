package cnabprovider

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/config"
	instancestorage "get.porter.sh/porter/pkg/instance-storage"
	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/stretchr/testify/require"
)

func Test_Install_Params(t *testing.T) {
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
				Name:          "mybuns",
				Version:       "v1.0.0",
				SchemaVersion: "v1.0.0",
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
				param.ApplyTo = []string{"not-install"}
				bun.Parameters["my-param"] = param
			}

			bytes, err := json.Marshal(bun)
			require.NoError(t, err)

			// We currently need to read/write from the same file on disk
			// as cnab-go's bundle loader still makes raw os calls for loading a bundle
			err = ioutil.WriteFile("testdata/bundle.json", bytes, 0644)
			require.NoError(t, err)

			args := ActionArguments{
				Claim:      "test",
				BundlePath: "testdata/bundle.json",
				Insecure:   true,
				Driver:     "debug",
			}
			if tc.provided {
				args.Params = map[string]string{
					"my-param": "my-param-value",
				}
			}

			err = d.Install(args)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				updatedClaim, err := d.instanceStorage.Read("test")
				require.NoError(t, err)
				require.Equal(t, tc.expectedVal, updatedClaim.Parameters["my-param"])
				require.NoError(t, err)
			}
		})
	}
}
