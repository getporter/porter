package kubernetes

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/require"

	yaml "gopkg.in/yaml.v2"
)

type UnInstallTest struct {
	expectedCommand string
	uninstallStep   UninstallStep
}

func TestMixin_UninstallStep(t *testing.T) {

	manifestDirectory := "/cnab/app/manifesto"

	deleteCmd := "kubectl delete -f"

	dontWait := false

	namespace := "meditations"

	selector := "app=nginx"
	forceIt := true
	withGrace := 1

	timeout := 1

	uninstallTests := []UnInstallTest{
		{
			expectedCommand: fmt.Sprintf("%s %s --wait", deleteCmd, manifestDirectory),
			uninstallStep: UninstallStep{
				UninstallArguments: &UninstallArguments{
					Manifests: []string{manifestDirectory},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait", deleteCmd, defaultManifestPath),
			uninstallStep: UninstallStep{
				UninstallArguments: &UninstallArguments{},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s", deleteCmd, defaultManifestPath),
			uninstallStep: UninstallStep{
				UninstallArguments: &UninstallArguments{
					Wait: &dontWait,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s", deleteCmd, defaultManifestPath, namespace),
			uninstallStep: UninstallStep{
				UninstallArguments: &UninstallArguments{
					Namespace: namespace,
					Wait:      &dontWait,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --selector=%s --wait", deleteCmd, defaultManifestPath, selector),
			uninstallStep: UninstallStep{
				UninstallArguments: &UninstallArguments{
					Selector: selector,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --force --grace-period=0 --wait", deleteCmd, defaultManifestPath),
			uninstallStep: UninstallStep{
				UninstallArguments: &UninstallArguments{
					Force: &forceIt,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --grace-period=%d --wait", deleteCmd, defaultManifestPath, withGrace),
			uninstallStep: UninstallStep{
				UninstallArguments: &UninstallArguments{
					GracePeriod: &withGrace,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --timeout=%ds --wait", deleteCmd, defaultManifestPath, timeout),
			uninstallStep: UninstallStep{
				UninstallArguments: &UninstallArguments{
					Timeout: &timeout,
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, uninstallTest := range uninstallTests {
		t.Run(uninstallTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, uninstallTest.expectedCommand)

			action := UninstallAction{Steps: []UninstallStep{uninstallTest.uninstallStep}}
			b, _ := yaml.Marshal(action)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err := h.UnInstall()

			require.NoError(t, err)
		})
	}
}
