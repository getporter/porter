package kubernetes

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/require"

	yaml "gopkg.in/yaml.v2"
)

type ExecuteTest struct {
	expectedCommand string
	executeStep     ExecuteStep
}

func TestMixin_ExecuteStep(t *testing.T) {

	manifestDirectory := "/cnab/app/manifests"

	upgradeCmd := "kubectl apply -f"

	dontWait := false

	recordIt := true
	validateIt := false

	namespace := "meditations"

	selector := "app=nginx"

	context := "context"

	forceIt := true
	withGrace := 1

	overwriteIt := false
	pruneIt := true

	timeout := 1

	upgradeTests := []ExecuteTest{
		// These tests are largely the same as the install, just testing that the embedded
		// install gets handled correctly
		{
			expectedCommand: fmt.Sprintf("%s %s --wait", upgradeCmd, manifestDirectory),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait", upgradeCmd, manifestDirectory),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s", upgradeCmd, manifestDirectory),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
						Wait:      &dontWait,
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s", upgradeCmd, manifestDirectory, namespace),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
						Namespace: namespace,
						Wait:      &dontWait,
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s --validate=false", upgradeCmd, manifestDirectory, namespace),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
						Namespace: namespace,
						Validate:  &validateIt,
						Wait:      &dontWait,
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s --record=true", upgradeCmd, manifestDirectory, namespace),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
						Namespace: namespace,
						Record:    &recordIt,
						Wait:      &dontWait,
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --selector=%s --wait", upgradeCmd, manifestDirectory, selector),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
						Selector:  selector,
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --context=%s --wait", upgradeCmd, manifestDirectory, context),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
						Context:  context,
					},
				},
			},
		},

		// These tests exercise the upgrade options
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --force --grace-period=0", upgradeCmd, manifestDirectory),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					Force: &forceIt,
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --grace-period=%d", upgradeCmd, manifestDirectory, withGrace),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					GracePeriod: &withGrace,
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "Hello",
						},
						Manifests: []string{manifestDirectory},
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --overwrite=false", upgradeCmd, manifestDirectory),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					Overwrite: &overwriteIt,
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "upgrade",
						},
						Manifests: []string{manifestDirectory},
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --prune=true", upgradeCmd, manifestDirectory),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					Prune: &pruneIt,
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "upgrade",
						},
						Manifests: []string{manifestDirectory},
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --timeout=%ds", upgradeCmd, manifestDirectory, timeout),
			executeStep: ExecuteStep{
				ExecuteInstruction: ExecuteInstruction{
					Timeout: &timeout,
					InstallArguments: InstallArguments{
						Step: Step{
							Description: "upgrade",
						},
						Manifests: []string{manifestDirectory},
					},
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, upgradeTest := range upgradeTests {
		t.Run(upgradeTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, upgradeTest.expectedCommand)

			action := ExecuteAction{Steps: []ExecuteStep{upgradeTest.executeStep}}
			b, _ := yaml.Marshal(action)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err := h.Execute()

			require.NoError(t, err)
		})
	}
}
