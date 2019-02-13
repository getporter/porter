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
	c.SetupPorterHome()

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	assert.Equal(t, "/root/.porter", home)
}

func TestConfig_GetHomeDirFromSymlink(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	// Set up no PORTER_HOME, and /usr/local/bin/porter -> ~/.porter/porter
	os.Unsetenv(EnvHOME)
	getExecutable = func() (string, error) {
		return "/usr/local/bin/porter", nil
	}
	evalSymlinks = func(path string) (string, error) {
		return "/root/.porter/porter", nil
	}

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	// The reason why we do filepath.join here and not above is because resolving symlinks gets the OS involved
	// and on Windows, that means flipping the afero `/` to `\`.
	assert.Equal(t, filepath.Join("/root", ".porter"), home)
}

func TestConfig_GetPorterConfigTemplate(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	gotTmpl, err := c.GetPorterConfigTemplate()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/porter.yaml")
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestConfig_GetRunScriptTemplate(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	gotTmpl, err := c.GetRunScriptTemplate()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/run")
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
