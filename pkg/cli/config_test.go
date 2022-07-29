package cli

import (
	"context"
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
		cmd.Flags().StringVar(&c.Data.Verbosity, "verbosity", "info", "verbosity")
		cmd.Flags().StringVar(&buildDriver, "driver", "", "build driver")
		cmd.Flag("driver").Annotations = map[string][]string{
			"viper-key": {"build-driver"},
		}

		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			return c.Load(context.Background(), nil)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
		c.DataLoader = LoadHierarchicalConfig(cmd)
		return cmd
	}

	t.Run("no flag", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/home/myuser/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()
		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, "info", c.Data.Verbosity, "config.Verbosity was not set correctly")
	})

	t.Run("verbosity flag", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/home/myuser/.porter")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--verbosity=warn"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, "warn", c.Data.Verbosity, "config.Verbosity was not set correctly")
	})

	t.Run("verbosity flag overrides config", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/home/myuser/.porter")
		c.TestContext.AddTestFileFromRoot("pkg/config/testdata/config.toml", "/home/myuser/.porter/config.toml")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--verbosity=error"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, "error", c.Data.Verbosity, "config.Verbosity was not set correctly")
	})

	t.Run("verbosity env var", func(t *testing.T) {
		os.Setenv("PORTER_VERBOSITY", "debug")
		defer os.Unsetenv("PORTER_VERBOSITY")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/home/myuser/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, "debug", c.Data.Verbosity, "config.Verbosity was not set correctly")
	})

	t.Run("build-driver env var", func(t *testing.T) {
		os.Setenv("PORTER_BUILD_DRIVER", config.BuildDriverBuildkit)
		defer os.Unsetenv("PORTER_BUILD_DRIVER")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/home/myuser/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, config.BuildDriverBuildkit, c.Data.BuildDriver, "c.Data.BuildDriver was not set correctly")
	})

	t.Run("build-driver from config", func(t *testing.T) {
		os.Unsetenv("PORTER_BUILD_DRIVER")
		defer os.Unsetenv("PORTER_BUILD_DRIVER")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/home/myuser/.porter")
		c.TestContext.AddTestFileFromRoot("pkg/config/testdata/config.toml", "/home/myuser/.porter/config.toml")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, config.BuildDriverBuildkit, c.Data.BuildDriver, "c.Data.BuildDriver was not set correctly")
	})

	t.Run("invalid verbosity env var", func(t *testing.T) {
		os.Setenv("PORTER_VERBOSITY", "blorp")
		defer os.Unsetenv("PORTER_VERBOSITY")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/home/myuser/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, config.LogLevel("info"), c.GetVerbosity(), "config.Verbosity was not set correctly")
	})

	t.Run("debug env var overrides config", func(t *testing.T) {
		os.Setenv("PORTER_VERBOSITY", "error")
		defer os.Unsetenv("PORTER_VERBOSITY")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/home/myuser/.porter")
		c.TestContext.AddTestFileFromRoot("pkg/config/testdata/config.toml", "/home/myuser/.porter/config.toml")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, "error", c.Data.Verbosity, "config.Verbosity should have been set by the env var and not the config")
	})

	t.Run("flag overrides debug env var overrides config", func(t *testing.T) {
		os.Setenv("PORTER_VERBOSITY", "warn")
		defer os.Unsetenv("PORTER_VERBOSITY")

		c := config.NewTestConfig(t)
		c.SetHomeDir("/home/myuser/.porter")
		c.TestContext.AddTestFileFromRoot("pkg/config/testdata/config.toml", "/home/myuser/.porter/config.toml")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--verbosity", "debug"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		assert.Equal(t, "debug", c.Data.Verbosity, "config.Verbosity should have been set by the flag and not the env var or config")
	})
}
