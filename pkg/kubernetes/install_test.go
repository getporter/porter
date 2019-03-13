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

type InstallTest struct {
	expectedCommand string
	installStep     InstallStep
}

func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}

func TestMixin_InstallStep(t *testing.T) {

	manifestDirectory := "/cnab/app/manifesto"

	installCmd := "kubectl apply -f"

	dontWait := false
	recordIt := true
	validateIt := false

	namespace := "meditations"

	selector := "app=nginx"

	installTests := []InstallTest{
		{
			expectedCommand: fmt.Sprintf("%s %s --wait", installCmd, manifestDirectory),
			installStep: InstallStep{
				InstallArguments: &InstallArguments{
					Manifests: []string{manifestDirectory},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait", installCmd, defaultManifestPath),
			installStep: InstallStep{
				InstallArguments: &InstallArguments{},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s", installCmd, defaultManifestPath),
			installStep: InstallStep{
				InstallArguments: &InstallArguments{
					Wait: &dontWait,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s", installCmd, defaultManifestPath, namespace),
			installStep: InstallStep{
				InstallArguments: &InstallArguments{
					Namespace: namespace,
					Wait:      &dontWait,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s --validate=false", installCmd, defaultManifestPath, namespace),
			installStep: InstallStep{
				InstallArguments: &InstallArguments{
					Namespace: namespace,
					Validate:  &validateIt,
					Wait:      &dontWait,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s --record=true", installCmd, defaultManifestPath, namespace),
			installStep: InstallStep{
				InstallArguments: &InstallArguments{
					Namespace: namespace,
					Record:    &recordIt,
					Wait:      &dontWait,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --selector=%s --wait", installCmd, defaultManifestPath, selector),
			installStep: InstallStep{
				InstallArguments: &InstallArguments{
					Selector: selector,
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, installTest := range installTests {
		t.Run(installTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, installTest.expectedCommand)

			action := InstallAction{Steps: []InstallStep{installTest.installStep}}
			b, _ := yaml.Marshal(action)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err := h.Install()

			require.NoError(t, err)
		})
	}
}
