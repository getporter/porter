package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	p := NewTestPorter(t)

	err := p.Create()
	require.NoError(t, err)

	configFileExists, err := p.FileSystem.Exists("porter.yaml")
	require.NoError(t, err)
	assert.True(t, configFileExists)

	// Verify that helpers is present and executable
	helperFileStats, err := p.FileSystem.Stat("helpers.sh")
	require.NoError(t, err)
	assert.Equal(t, "-rwxr-xr-x", helperFileStats.Mode().String())

	dockerfileExists, err := p.FileSystem.Exists("Dockerfile.tmpl")
	require.NoError(t, err)
	assert.True(t, dockerfileExists)

	readmeExists, err := p.FileSystem.Exists("README.md")
	require.NoError(t, err)
	assert.True(t, readmeExists)

	gitignore, err := p.FileSystem.Exists(".gitignore")
	require.NoError(t, err)
	assert.True(t, gitignore)

	dockerignore, err := p.FileSystem.Exists(".dockerignore")
	require.NoError(t, err)
	assert.True(t, dockerignore)
}
