package mixin

import (
	"testing"

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
