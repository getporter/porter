package main

import (
	"bytes"
	"strings"
	"testing"

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
