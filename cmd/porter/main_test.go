package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandWiring(t *testing.T) {
	testcases := []string{
		"build",
		"create",
		"install",
		"uninstall",
		"run",
		"schema",
		"bundles",
		"bundle create",
		"bundle build",
		"installation install",
		"installation uninstall",
		"mixins",
		"mixins list",
		"plugins list",
		"storage",
		"storage migrate",
		"version",
	}

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			osargs := strings.Split(tc, " ")

			rootCmd := buildRootCommand()
			cmd, _, err := rootCmd.Find(osargs)
			assert.NoError(t, err)
			assert.Equal(t, osargs[len(osargs)-1], cmd.Name())
		})
	}
}

func TestHelp(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		var output bytes.Buffer
		rootCmd := buildRootCommand()
		rootCmd.SetArgs([]string{})
		rootCmd.SetOut(&output)

		err := rootCmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, output.String(), "Usage")
	})

	t.Run("help", func(t *testing.T) {
		var output bytes.Buffer
		rootCmd := buildRootCommand()
		rootCmd.SetArgs([]string{"help"})
		rootCmd.SetOut(&output)

		err := rootCmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, output.String(), "Usage")
	})

	t.Run("--help", func(t *testing.T) {
		var output bytes.Buffer
		rootCmd := buildRootCommand()
		rootCmd.SetArgs([]string{"--help"})
		rootCmd.SetOut(&output)

		err := rootCmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, output.String(), "Usage")
	})
}

func TestExperimentalFlags(t *testing.T) {
	// do not run in parallel
	expEnvVar := "PORTER_EXPERIMENTAL"
	os.Unsetenv(expEnvVar)

	t.Run("flag unset, env unset", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		cmd.Execute()
		assert.False(t, p.Config.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("flag set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--experimental", experimental.NoopFeature})
		cmd.Execute()

		assert.True(t, p.Config.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("flag unset, env set", func(t *testing.T) {
		os.Setenv(expEnvVar, experimental.NoopFeature)
		defer os.Unsetenv(expEnvVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		cmd.Execute()

		assert.True(t, p.Config.IsFeatureEnabled(experimental.FlagNoopFeature))
	})
}
