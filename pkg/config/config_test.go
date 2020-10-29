package config

import (
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
