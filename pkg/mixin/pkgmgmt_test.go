package mixin

import (
	"os"
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
		wantCommand   string
	}{
		{"build", "build", "/root/.porter/mixins/exec/exec build"},
		{"install", "install", "/root/.porter/mixins/exec/exec install"},
		{"upgrade", "upgrade", "/root/.porter/mixins/exec/exec upgrade"},
		{"uninstall", "uninstall", "/root/.porter/mixins/exec/exec uninstall"},
		{"invoke", "status", "/root/.porter/mixins/exec/exec invoke --action status"},
		{"version", "version --output json", "/root/.porter/mixins/exec/exec version --output json"},
	}

	os.Unsetenv(test.ExpectedCommandEnv)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			r := client.NewTestRunner(t, "exec", "mixins", false)
			r.Debug = false
			os.Setenv(test.ExpectedCommandEnv, tc.wantCommand)

			mgr := PackageManager{}
			cmd := pkgmgmt.CommandOptions{Command: tc.runnerCommand, PreRun: mgr.PreRunMixinCommandHandler}
			err := r.Run(cmd)
			require.NoError(t, err)
		})
	}
}
