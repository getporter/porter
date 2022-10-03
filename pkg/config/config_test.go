package config

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_GetHomeDir(t *testing.T) {
	c := NewTestConfig(t)

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	assert.Equal(t, "/home/myuser/.porter", home)
}

func TestConfig_GetHomeDirFromSymlink(t *testing.T) {
	c := NewTestConfig(t)

	// Set up no PORTER_HOME, and /usr/local/bin/porter -> ~/.porter/porter
	c.Unsetenv(EnvHOME)
	getExecutable = func() (string, error) {
		return "/usr/local/bin/porter", nil
	}
	evalSymlinks = func(path string) (string, error) {
		return "/home/myuser/.porter/porter", nil
	}

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	// The reason why we do filepath.join here and not above is because resolving symlinks gets the OS involved
	// and on Windows, that means flipping the afero `/` to `\`.
	assert.Equal(t, filepath.Join("/home/myuser", ".porter"), home)
}

func TestConfig_GetFeatureFlags(t *testing.T) {
	t.Parallel()

	t.Run("feature defaulted to disabled", func(t *testing.T) {
		c := Config{}
		assert.False(t, c.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("feature enabled", func(t *testing.T) {
		c := Config{}
		c.Data.ExperimentalFlags = []string{experimental.NoopFeature}
		assert.True(t, c.IsFeatureEnabled(experimental.FlagNoopFeature))
	})
}

func TestConfigExperimentalFlags(t *testing.T) {
	// Do not run in parallel since we are using os.Setenv

	t.Run("default off", func(t *testing.T) {
		c := NewTestConfig(t)
		assert.False(t, c.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("user configuration", func(t *testing.T) {
		os.Setenv("PORTER_EXPERIMENTAL", experimental.NoopFeature)
		defer os.Unsetenv("PORTER_EXPERIMENTAL")

		c := New()
		require.NoError(t, c.Load(context.Background(), nil), "Load failed")
		assert.True(t, c.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("programmatically enabled", func(t *testing.T) {
		c := NewTestConfig(t)
		c.Data.ExperimentalFlags = nil
		c.SetExperimentalFlags(experimental.FlagNoopFeature)
		assert.True(t, c.IsFeatureEnabled(experimental.FlagNoopFeature))
	})
}

func TestConfig_GetBuildDriver(t *testing.T) {
	c := NewTestConfig(t)
	c.Data.BuildDriver = "special"
	require.Equal(t, BuildDriverBuildkit, c.GetBuildDriver(), "Default to docker when experimental is false, even when a build driver is set")
}

func TestConfig_ExportRemoteConfigAsEnvironmentVariables(t *testing.T) {
	ctx := context.Background()

	c := NewTestConfig(t)
	c.DataLoader = LoadFromEnvironment()
	c.TestContext.AddTestFile("testdata/config.toml", "/home/myuser/.porter/config.toml")

	err := c.Load(ctx, nil)
	require.NoError(t, err, "Config.Load failed")

	gotEnvVars := c.ExportRemoteConfigAsEnvironmentVariables()
	sort.Strings(gotEnvVars)
	wantEnvVars := []string{
		"PORTER_LOGS_LEVEL=info",
		"PORTER_LOGS_LOG_TO_FILE=true",
		"PORTER_LOGS_STRUCTURED=true",
		"PORTER_TELEMETRY_ENABLED=true",
		"PORTER_TELEMETRY_REDIRECT_TO_FILE=true",
		"PORTER_VERBOSITY=warn",
	}
	assert.Equal(t, wantEnvVars, gotEnvVars)
}
