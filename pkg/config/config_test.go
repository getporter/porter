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
	c := NewTestConfig(t)

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
	c := NewTestConfig(t)
	c.SetupPorterHome()

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	assert.Equal(t, "/root/.porter", home)
}

func TestConfig_GetPorterConfigTemplate(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	gotTmpl, err := c.GetPorterConfigTemplate()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("../../templates/porter.yaml")
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestConfig_GetRunScriptTemplate(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	gotTmpl, err := c.GetRunScriptTemplate()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("../../templates/run")
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestConfig_GetBundleDir(t *testing.T) {
	c := NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)
	c.TestContext.AddTestDirectory("testdata/bundles", "bundles")

	err := c.LoadManifest()
	require.NoError(t, err)

	result, err := c.GetBundleDir("mysql")
	require.NoError(t, err)
	assert.Equal(t, "bundles/mysql", result)
}

func TestConfig_GetBundleDir_BundleNotInstalled(t *testing.T) {
	c := NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/missingdep.porter.yaml", Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	result, err := c.GetBundleDir("mysql")
	require.NoError(t, err)
	assert.Equal(t, "bundles/mysql", result)
}
