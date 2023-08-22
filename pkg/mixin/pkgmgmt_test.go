package mixin

import (
	"context"
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
		{"build", "build", "/home/myuser/.porter/mixins/exec/exec build\n\\home\\myuser\\.porter\\mixins\\exec\\exec.exe build"},
		{"install", "install", "/home/myuser/.porter/mixins/exec/exec install\n\\home\\myuser\\.porter\\mixins\\exec\\exec.exe install"},
		{"upgrade", "upgrade", "/home/myuser/.porter/mixins/exec/exec upgrade\n\\home\\myuser\\.porter\\mixins\\exec\\exec.exe upgrade"},
		{"uninstall", "uninstall", "/home/myuser/.porter/mixins/exec/exec uninstall\n\\home\\myuser\\.porter\\mixins\\exec\\exec.exe uninstall"},
		{"invoke", "status", "/home/myuser/.porter/mixins/exec/exec invoke --action status\n\\home\\myuser\\.porter\\mixins\\exec\\exec.exe invoke --action status"},
		{"version", "version --output json", "/home/myuser/.porter/mixins/exec/exec version --output json\n\\home\\myuser\\.porter\\mixins\\exec\\exec.exe version --output json"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			r := client.NewTestRunner(t, "exec", "mixins", false)
			r.Setenv(test.ExpectedCommandEnv, tc.wantCommand)

			mgr := PackageManager{}
			cmd := pkgmgmt.CommandOptions{Command: tc.runnerCommand, PreRun: mgr.PreRunMixinCommandHandler}
			err := r.Run(ctx, cmd)
			require.NoError(t, err)
		})
	}
}
