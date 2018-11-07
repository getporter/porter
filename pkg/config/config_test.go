package config

import (
	"io/ioutil"
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

	// Set up a test porter home
	testHome := "/root/.porter"
	err := c.FileSystem.MkdirAll(testHome, os.ModePerm)
	require.NoError(t, err)
	os.Setenv(EnvHOME, testHome)

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	assert.Equal(t, testHome, home)
}

func TestConfig_GetPorterConfigTemplate(t *testing.T) {
	c, _ := NewTestConfig()

	// Set up a test porter home
	testHome := "/root/.porter"
	err := c.FileSystem.MkdirAll(testHome, os.ModePerm)
	require.NoError(t, err)
	os.Setenv(EnvHOME, testHome)

	// Setup a templates directory
	templatesDir, err := c.GetTemplatesDir()
	require.NoError(t, err)
	err = c.FileSystem.Mkdir(templatesDir, os.ModePerm)
	require.NoError(t, err)

	// Add a template porter.yaml
	tmpl, err := ioutil.ReadFile(filepath.Join("testdata/porter.yaml"))
	require.NoError(t, err)
	err = c.FileSystem.WriteFile(filepath.Join(templatesDir, Name), tmpl, os.ModePerm)
	require.NoError(t, err)

	gotTmpl, err := c.GetPorterConfigTemplate()
	require.NoError(t, err)

	assert.Equal(t, tmpl, gotTmpl)
}
