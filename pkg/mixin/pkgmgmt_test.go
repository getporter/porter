package mixin

import (
	"fmt"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/client"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/require"
)

func TestRunner_BuildCommand(t *testing.T) {
	testcases := []struct {
		name          string
		runnerCommand string
		wantMixin     string
		wantArgs      string
	}{
		{"build", "build", "exec", "build"},
		{"install", "install", "exec", "install"},
		{"upgrade", "upgrade", "exec", "upgrade"},
		{"uninstall", "uninstall", "exec", "uninstall"},
		{"invoke", "status", "exec", "invoke --action status"},
		{"version", "version --output json", "exec", "version --output json"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			r := client.NewTestRunner(t, "exec", "mixins", false)
			r.Debug = false
			wantCommand := fmt.Sprintf("%s %s", filepath.Join(r.TestConfig.GetMixinsDir(), tc.wantMixin, tc.wantMixin+pkgmgmt.FileExt), tc.wantArgs)
			r.Setenv(test.ExpectedCommandEnv, wantCommand)

			mgr := PackageManager{}
			cmd := pkgmgmt.CommandOptions{Command: tc.runnerCommand, PreRun: mgr.PreRunMixinCommandHandler}
			err := r.Run(cmd)
			require.NoError(t, err)
		})
	}
}
