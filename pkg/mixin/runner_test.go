package mixin

import (
	"bytes"
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

func TestRunner_Run(t *testing.T) {
	output := &bytes.Buffer{}

	// I'm not using the TestRunner because I want to use the current filesystem, not an isolated one
	r := NewRunner("exec", "../../bin/mixins/exec", false)
	r.Command = "install"
	r.File = "testdata/exec_input.yaml"

	// Capture the output
	r.Out = output
	r.Err = output

	err := r.Validate()
	require.NoError(t, err)

	err = r.Run()
	assert.NoError(t, err)
	assert.Contains(t, string(output.Bytes()), "Hello World")
}
