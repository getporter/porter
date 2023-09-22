package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
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

// Validate that porter is correctly binding experimental which is a flag on some commands AND is persisted on config.Data
// This is a regression test to ensure that we are applying our configuration from viper and cobra in the proper order
// such that flags defined on config.Data are persisted.
// I'm testing both experimental and verbosity because honestly, I've seen both break enough that I'd rather have excessive test than see it break again.
func TestExperimentalFlags(t *testing.T) {
	// do not run in parallel
	expEnvVar := "PORTER_EXPERIMENTAL"
	os.Unsetenv(expEnvVar)

	t.Run("default", func(t *testing.T) {
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

	t.Run("env set", func(t *testing.T) {
		os.Setenv(expEnvVar, experimental.NoopFeature)
		defer os.Unsetenv(expEnvVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		cmd.Execute()

		assert.True(t, p.Config.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`experimental: [no-op]`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		cmd.Execute()

		assert.True(t, p.Config.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("flag set, cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`experimental: []`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--experimental", "no-op"})
		cmd.Execute()

		assert.True(t, p.Config.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("flag set, env set", func(t *testing.T) {
		os.Setenv(expEnvVar, "")
		defer os.Unsetenv(expEnvVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--experimental", "no-op"})
		cmd.Execute()

		assert.True(t, p.Config.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("env set, cfg set", func(t *testing.T) {
		os.Setenv(expEnvVar, "")
		defer os.Unsetenv(expEnvVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`experimental: [no-op]`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		cmd.Execute()

		assert.False(t, p.Config.IsFeatureEnabled(experimental.FlagNoopFeature))
	})
}

// Validate that porter is correctly binding verbosity which is a flag on all commands AND is persisted on config.Data
// This is a regression test to ensure that we are applying our configuration from viper and cobra in the proper order
// such that flags defined on config.Data are persisted.
// I'm testing both experimental and verbosity because honestly, I've seen both break enough that I'd rather have excessive test than see it break again.
func TestVerbosity(t *testing.T) {
	// do not run in parallel
	envVar := "PORTER_VERBOSITY"
	os.Unsetenv(envVar)

	t.Run("default", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		cmd.Execute()
		assert.Equal(t, config.LogLevelInfo, p.Config.GetVerbosity())
	})

	t.Run("flag set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--verbosity=debug"})
		cmd.Execute()

		assert.Equal(t, config.LogLevelDebug, p.Config.GetVerbosity())
	})

	t.Run("env set", func(t *testing.T) {
		os.Setenv(envVar, "error")
		defer os.Unsetenv(envVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		cmd.Execute()

		assert.Equal(t, config.LogLevelError, p.Config.GetVerbosity())
	})

	t.Run("cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`verbosity: warning`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		cmd.Execute()

		assert.Equal(t, config.LogLevelWarn, p.Config.GetVerbosity())
	})

	t.Run("flag set, cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`verbosity: debug`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--verbosity", "warn"})
		cmd.Execute()

		assert.Equal(t, config.LogLevelWarn, p.Config.GetVerbosity())
	})

	t.Run("flag set, env set", func(t *testing.T) {
		os.Setenv(envVar, "warn")
		defer os.Unsetenv(envVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--verbosity=debug"})
		cmd.Execute()

		assert.Equal(t, config.LogLevelDebug, p.Config.GetVerbosity())
	})

	t.Run("env set, cfg set", func(t *testing.T) {
		os.Setenv(envVar, "warn")
		defer os.Unsetenv(envVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`verbosity: debug`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		cmd.Execute()

		assert.Equal(t, config.LogLevelWarn, p.Config.GetVerbosity())
	})
}

// Validate that porter is correctly binding porter explain --output which is a flag that is NOT bound to config.Data
// This is a regression test to ensure that we are applying our configuration from viper and cobra in the proper order
// such that flags defined on a separate data structure from config.Data are persisted.
func TestExplainOutput(t *testing.T) {
	// do not run in parallel
	envVar := "PORTER_OUTPUT"
	os.Unsetenv(envVar)

	const ref = "ghcr.io/getporter/examples/porter-hello:v0.2.0"

	assertPlainOutput := func(t *testing.T, output string) {
		t.Helper()
		assert.Contains(t, "Name: examples/porter-hello", output, "explain should have output plain text")
	}

	assertJsonOutput := func(t *testing.T, output string) {
		t.Helper()
		assert.Contains(t, `"name": "examples/porter-hello",`, output, "explain should have output JSON")
	}

	assertYamlOutput := func(t *testing.T, output string) {
		t.Helper()
		assert.Contains(t, `- name: name`, output, "explain should have output YAML")
	}

	t.Run("default", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"explain", ref})
		require.NoError(t, cmd.Execute(), "explain failed")

		assertPlainOutput(t, p.TestConfig.TestContext.GetOutput())
	})

	t.Run("flag set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"explain", ref, "--output=json"})
		require.NoError(t, cmd.Execute(), "explain failed")

		assertJsonOutput(t, p.TestConfig.TestContext.GetOutput())
	})

	t.Run("env set", func(t *testing.T) {
		os.Setenv(envVar, "json")
		defer os.Unsetenv(envVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"explain", ref})
		require.NoError(t, cmd.Execute(), "explain failed")

		assertJsonOutput(t, p.TestConfig.TestContext.GetOutput())
	})

	t.Run("cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`output: json`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"explain", ref})
		require.NoError(t, cmd.Execute(), "explain failed")

		assertJsonOutput(t, p.TestConfig.TestContext.GetOutput())
	})

	t.Run("flag set, cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`output: json`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"explain", ref, "--output=yaml"})
		require.NoError(t, cmd.Execute(), "explain failed")

		assertYamlOutput(t, p.TestConfig.TestContext.GetOutput())
	})

	t.Run("flag set, env set", func(t *testing.T) {
		os.Setenv(envVar, "json")
		defer os.Unsetenv(envVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"explain", ref, "--output=yaml"})
		require.NoError(t, cmd.Execute(), "explain failed")

		assertYamlOutput(t, p.TestConfig.TestContext.GetOutput())
	})

	t.Run("env set, cfg set", func(t *testing.T) {
		os.Setenv(envVar, "yaml")
		defer os.Unsetenv(envVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`output: json`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"explain", ref})
		require.NoError(t, cmd.Execute(), "explain failed")

		assertYamlOutput(t, p.TestConfig.TestContext.GetOutput())
	})
}
