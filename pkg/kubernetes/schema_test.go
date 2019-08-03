package kubernetes

import (
	//"bytes"
	//"fmt"
	"io/ioutil"
	//"os"
	"testing"

	//"github.com/deislabs/porter/pkg/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	//yaml "gopkg.in/yaml.v2"
)

func TestMixin_PrintSchema(t *testing.T) {
	m := NewTestMixin(t)

	err := m.PrintSchema()
	require.NoError(t, err)

	gotSchema := m.TestContext.GetOutput()

	wantSchema, err := ioutil.ReadFile("testdata/schema.json")
	require.NoError(t, err)

	assert.Equal(t, string(wantSchema), gotSchema)
}

func TestMixin_ValidatePayload(t *testing.T) {
	testcases := []struct {
		name  string
		step  string
		pass  bool
		error string
	}{
		{"install", "testdata/install-input.yaml", true, ""},
		{"upgrade", "testdata/upgrade-input.yaml", true, ""},
		{"invoke", "testdata/invoke-input.yaml", true, ""},
		{"uninstall", "testdata/uninstall-input.yaml", true, ""},
		{"install-bad-wait-flag", "testdata/install-input-bad-wait-flag.yaml", false, "install.0.kubernetes.wait: Invalid type. Expected: boolean, given: string"},
		{"install-no-manifests", "testdata/install-input-no-manifests.yaml", false, "install.0.kubernetes: manifests is required"},
		{"install-bad-outputs", "testdata/install-input-bad-outputs.yaml", false, "install.0.kubernetes.outputs.0: resourceType is required\n\t* install.0.kubernetes.outputs.0: resourceName is required\n\t* install.0.kubernetes.outputs.0: jsonPath is required"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewTestMixin(t)
			b, err := ioutil.ReadFile(tc.step)
			require.NoError(t, err)

			err = m.ValidatePayload(b)
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tc.error)
			}
		})
	}
}

// TODO: Currently can't test this with pkg/test because it runs multiple commands
// func TestOuputSchema(t *testing.T) {
// 	manifestDirectory := "/cnab/app/manifests"

// 	installCmd := "kubectl apply -f"

// 	installTests := []InstallTest{
// 		{
// 			expectedCommand: fmt.Sprintf("%s %s --wait", installCmd, manifestDirectory),
// 			installStep: InstallStep{

// 				InstallArguments: InstallArguments{
// 					Step: Step{
// 						Description: "Hello",
// 						Outputs: []KubernetesOutput{
// 							KubernetesOutput{
// 								Name:         "test",
// 								Namespace:    "Default",
// 								ResourceType: "service",
// 								ResourceName: "aservice",
// 								JSONPath:     "a.path.",
// 							},
// 						},
// 					},
// 					Manifests: []string{manifestDirectory},
// 				},
// 			},
// 		},
// 		{
// 			expectedCommand: fmt.Sprintf("%s %s --wait", installCmd, manifestDirectory),
// 			installStep: InstallStep{

// 				InstallArguments: InstallArguments{
// 					Step: Step{
// 						Description: "Hello",
// 						Outputs: []KubernetesOutput{
// 							KubernetesOutput{
// 								Name:         "test",
// 								ResourceType: "service",
// 								ResourceName: "aservice",
// 								JSONPath:     "a.path.",
// 							},
// 						},
// 					},
// 					Manifests: []string{manifestDirectory},
// 				},
// 			},
// 		},
// 	}

// 	defer os.Unsetenv(test.ExpectedCommandEnv)
// 	for _, installTest := range installTests {
// 		t.Run(installTest.expectedCommand, func(t *testing.T) {
// 			os.Setenv(test.ExpectedCommandEnv, installTest.expectedCommand)

// 			action := InstallAction{Steps: []InstallStep{installTest.installStep}}
// 			b, _ := yaml.Marshal(action)

// 			h := NewTestMixin(t)
// 			h.In = bytes.NewReader(b)

// 			err := h.Install()

// 			require.NoError(t, err)
// 		})
// 	}
// }
