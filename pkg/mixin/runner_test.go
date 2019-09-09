package mixin

import (
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_Validate(t *testing.T) {
	r := NewTestRunner(t, "exec", true)

	r.File = "exec_input.yaml"
	r.TestContext.AddTestFile("testdata/exec_input.yaml", r.File)

	err := r.Validate()
	require.NoError(t, err)
}

func TestRunner_Validate_MissingName(t *testing.T) {
	// Setup failure: empty mixin name
	r := NewTestRunner(t, "", true)

	err := r.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mixin not specified")
}

func TestRunner_Validate_MissingExecutable(t *testing.T) {
	r := NewTestRunner(t, "exec", true)

	// Setup failure: Don't copy the mixin binary into the test context
	err := r.FileSystem.Remove(r.getMixinPath())
	require.NoError(t, err)

	err = r.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mixin not found")
}

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
			r := NewTestRunner(t, "exec", false)
			r.Debug = false
			r.Command = tc.runnerCommand
			os.Setenv(test.ExpectedCommandEnv, tc.wantCommand)

			err := r.Run()
			require.NoError(t, err)
		})
	}
}
