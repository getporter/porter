package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_Validate(t *testing.T) {
	r := NewTestRunner(t, "lucky-charms", "cereals", true)

	err := r.Validate()
	require.NoError(t, err)
}

func TestRunner_Validate_MissingName(t *testing.T) {
	// Setup failure: empty package name
	r := NewTestRunner(t, "", "candy", true)

	err := r.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "package name to execute not specified")
}

func TestRunner_Validate_MissingExecutable(t *testing.T) {
	r := NewTestRunner(t, "mypackage", "packages", true)

	// Setup failure: Don't copy the package binary into the test context
	err := r.FileSystem.Remove(r.getExecutablePath())
	require.NoError(t, err)

	err = r.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "package not found")
}
