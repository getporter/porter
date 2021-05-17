package cli

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadHierarchicalConfig(t *testing.T) {
	// Cannot be run in parallel because viper reads directly from env vars
	buildCommand := func(c *config.Config) *cobra.Command {
		var buildDriver string
		cmd := &cobra.Command{}
		cmd.Flags().BoolVar(&c.Debug, "debug", false, "debug")
		cmd.Flags().BoolVar(&c.DebugPlugins, "debug-plugins", false, "debug plugins")
		cmd.Flags().StringVar(&buildDriver, "driver", "", "build driver")
		cmd.Flag("driver").Annotations = map[string][]string{
			"viper-key": {"build-driver"},
		}

		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			return c.LoadData()
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
		c.DataLoader = LoadHierarchicalConfig(cmd)
		return cmd
	}

	t.Run("no flag", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()
		require.NoError(t, err, "dataloader failed")
		assert.False(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug flag", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--debug"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.True(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug flag overrides config", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		c.TestContext.AddTestFileFromRoot("pkg/config/testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--debug=false"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.False(t, c.Debug, "config.Debug should have been set by the flag and not the config")
	})

	t.Run("debug env var", func(t *testing.T) {
		os.Setenv("PORTER_DEBUG", "true")
		defer os.Unsetenv("PORTER_DEBUG")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.True(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug-plugins env var", func(t *testing.T) {
		os.Setenv("PORTER_DEBUG_PLUGINS", "true")
		defer os.Unsetenv("PORTER_DEBUG_PLUGINS")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.True(t, c.DebugPlugins, "config.DebugPlugins was not set correctly")
	})

	t.Run("build-driver env var", func(t *testing.T) {
		os.Setenv("PORTER_BUILD_DRIVER", config.BuildDriverBuildkit)
		defer os.Unsetenv("PORTER_BUILD_DRIVER")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, config.BuildDriverBuildkit, c.Data.BuildDriver, "c.Data.BuildDriver was not set correctly")
	})

	t.Run("build-driver from config", func(t *testing.T) {
		os.Unsetenv("PORTER_BUILD_DRIVER")
		defer os.Unsetenv("PORTER_EXPERIMENTAL")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		c.TestContext.AddTestFileFromRoot("pkg/config/testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, config.BuildDriverBuildkit, c.Data.BuildDriver, "c.Data.BuildDriver was not set correctly")
	})

	t.Run("invalid debug env var", func(t *testing.T) {
		os.Setenv("PORTER_DEBUG", "blorp")
		defer os.Unsetenv("PORTER_DEBUG")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.False(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug env var overrides config", func(t *testing.T) {
		os.Setenv("PORTER_DEBUG", "false")
		defer os.Unsetenv("PORTER_DEBUG")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		c.TestContext.AddTestFileFromRoot("pkg/config/testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.False(t, c.Debug, "config.Debug should have been set by the env var and not the config")
	})

	t.Run("flag overrides debug env var overrides config", func(t *testing.T) {
		os.Setenv("PORTER_DEBUG", "false")
		defer os.Unsetenv("PORTER_DEBUG")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		c.TestContext.AddTestFileFromRoot("pkg/config/testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--debug", "true"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.True(t, c.Debug, "config.Debug should have been set by the flag and not the env var or config")
	})
}
