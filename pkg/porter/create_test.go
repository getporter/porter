package porter

import (
	"os"
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

func TestCreateWithBuildkit(t *testing.T) {
	os.Setenv("PORTER_EXPERIMENTAL", "build-drivers")
	os.Setenv("PORTER_BUILD_DRIVER", "buildkit")
	defer os.Unsetenv("PORTER_EXPERIMENTAL")
	defer os.Unsetenv("PORTER_BUILD_DRIVER")

	p := NewTestPorter(t)

	p.LoadData()

	err := p.Create()
	require.NoError(t, err)

	dockerfile, err := p.FileSystem.ReadFile("Dockerfile.tmpl")
	require.NoError(t, err, "could not read template dockerfile")

	assert.Contains(t, string(dockerfile), "# syntax=docker/dockerfile:1.2")
}
