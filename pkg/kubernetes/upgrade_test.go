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

type UpgradeTest struct {
	expectedCommand string
	upgradeStep     UpgradeStep
}

func TestMixin_UpgradeStep(t *testing.T) {

	manifestDirectory := "/cnab/app/manifesto"

	upgradeCmd := "kubectl apply -f"

	dontWait := false

	recordIt := true
	validateIt := false

	namespace := "meditations"

	selector := "app=nginx"

	forceIt := true
	withGrace := 1

	overwriteIt := false
	pruneIt := true

	timeout := 1

	upgradeTests := []UpgradeTest{
		// These tests are largely the same as the install, just testing that the embedded
		// install gets handled correctly
		{
			expectedCommand: fmt.Sprintf("%s %s --wait", upgradeCmd, manifestDirectory),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					InstallArguments: &InstallArguments{
						Manifests: []string{manifestDirectory},
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait", upgradeCmd, defaultManifestPath),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					InstallArguments: &InstallArguments{},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s", upgradeCmd, defaultManifestPath),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					InstallArguments: &InstallArguments{
						Wait: &dontWait,
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s", upgradeCmd, defaultManifestPath, namespace),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					InstallArguments: &InstallArguments{
						Namespace: namespace,
						Wait:      &dontWait,
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s --validate=false", upgradeCmd, defaultManifestPath, namespace),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					InstallArguments: &InstallArguments{
						Namespace: namespace,
						Validate:  &validateIt,
						Wait:      &dontWait,
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s -n %s --record=true", upgradeCmd, defaultManifestPath, namespace),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					InstallArguments: &InstallArguments{
						Namespace: namespace,
						Record:    &recordIt,
						Wait:      &dontWait,
					},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --selector=%s --wait", upgradeCmd, defaultManifestPath, selector),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					InstallArguments: &InstallArguments{
						Selector: selector,
					},
				},
			},
		},

		// These tests exercise the upgrade options
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --force --grace-period=0", upgradeCmd, defaultManifestPath),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					Force:            &forceIt,
					InstallArguments: &InstallArguments{},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --grace-period=%d", upgradeCmd, defaultManifestPath, withGrace),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					GracePeriod:      &withGrace,
					InstallArguments: &InstallArguments{},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --overwrite=false", upgradeCmd, defaultManifestPath),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					Overwrite:        &overwriteIt,
					InstallArguments: &InstallArguments{},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --prune=true", upgradeCmd, defaultManifestPath),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					Prune:            &pruneIt,
					InstallArguments: &InstallArguments{},
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf("%s %s --wait --timeout=%ds", upgradeCmd, defaultManifestPath, timeout),
			upgradeStep: UpgradeStep{
				UpgradeArguments: &UpgradeArguments{
					Timeout:          &timeout,
					InstallArguments: &InstallArguments{},
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, upgradeTest := range upgradeTests {
		t.Run(upgradeTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, upgradeTest.expectedCommand)

			action := UpgradeAction{Steps: []UpgradeStep{upgradeTest.upgradeStep}}
			b, _ := yaml.Marshal(action)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err := h.Upgrade()

			require.NoError(t, err)
		})
	}
}
