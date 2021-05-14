package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
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
		"bundle install",
		"bundle uninstall",
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

func TestHelpText(t *testing.T) {
	rootCmd := buildRootCommand()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"help"})
	rootCmd.Execute()
	helpText := buf.String()
	assert.Contains(t, helpText, "Resources:")
	assert.Contains(t, helpText, "Aliased Commands:")
	assert.Contains(t, helpText, "Meta Commands:")
}

func TestExperimentalFlags(t *testing.T) {
	// do not run in parallel
	expEnvVar := "PORTER_EXPERIMENTAL"
	os.Unsetenv(expEnvVar)

	t.Run("flag unset, env unset", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{})
		cmd.Execute()
		assert.False(t, p.Config.IsFeatureEnabled(experimental.FlagBuildDrivers))
	})

	t.Run("flag set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"--experimental", experimental.BuildDrivers})
		cmd.Execute()

		assert.True(t, p.Config.IsFeatureEnabled(experimental.FlagBuildDrivers))
	})

	t.Run("flag unset, env set", func(t *testing.T) {
		os.Setenv(expEnvVar, experimental.BuildDrivers)
		defer os.Unsetenv(expEnvVar)

		p := porter.NewTestPorter(t)
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{})
		cmd.Execute()

		assert.True(t, p.Config.IsFeatureEnabled(experimental.FlagBuildDrivers))
	})
}
