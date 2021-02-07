package config

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_GetHomeDir(t *testing.T) {
	c := NewTestConfig(t)

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	assert.Equal(t, "/root/.porter", home)
}

func TestConfig_GetHomeDirFromSymlink(t *testing.T) {
	c := NewTestConfig(t)

	// Set up no PORTER_HOME, and /usr/local/bin/porter -> ~/.porter/porter
	c.Unsetenv(EnvHOME)
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

func TestConfig_GetFeatureFlags(t *testing.T) {
	t.Run("build builders defaulted to disabled", func(t *testing.T) {
		c := Config{
			Data: &Data{},
		}
		assert.False(t, c.IsFeatureEnabled(experimental.FlagBuildDrivers))
	})

	t.Run("build builders enabled", func(t *testing.T) {
		c := Config{
			Data: &Data{
				ExperimentalFlags: []string{experimental.BuildDrivers},
			},
		}
		assert.True(t, c.IsFeatureEnabled(experimental.FlagBuildDrivers))
	})
}
