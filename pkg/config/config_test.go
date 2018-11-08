package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestConfig_GetHomeDir(t *testing.T) {
	c, _ := NewTestConfig()

	// Set up a test porter home
	testEntrypoint, err := os.Executable()
	testHome := filepath.Dir(testEntrypoint)
	require.NoError(t, err)
	err = c.FileSystem.MkdirAll(testHome, os.ModePerm)
	require.NoError(t, err)

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	assert.Equal(t, testHome, home)
}

func TestConfig_GetHomeDir_HomeSet(t *testing.T) {
	c, _ := NewTestConfig()
	SetupPorterHome(t, c)

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	assert.Equal(t, "/root/.porter", home)
}

func TestConfig_GetPorterConfigTemplate(t *testing.T) {
	c, _ := NewTestConfig()
	templates := SetupPorterHome(t, c)

	gotTmpl, err := c.GetPorterConfigTemplate()
	require.NoError(t, err)

	assert.Equal(t, templates["porter.yaml"], gotTmpl)
}

func TestConfig_GetRunScriptTemplate(t *testing.T) {
	c, _ := NewTestConfig()
	templates := SetupPorterHome(t, c)

	gotTmpl, err := c.GetRunScriptTemplate()
	require.NoError(t, err)

	assert.Equal(t, templates["run"], gotTmpl)
}
