package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/spf13/pflag"
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

func TestShouldSkipSecrets(t *testing.T) {
	testcases := map[string]bool{
		"plugins install":      true,
		"plugins list":         true,
		"mixins install":       true,
		"mixins feed generate": true,
		"installation list":    false,
		"config show":          false,
	}

	for tc, want := range testcases {
		t.Run(tc, func(t *testing.T) {
			rootCmd := buildRootCommand()
			cmd, _, err := rootCmd.Find(strings.Split(tc, " "))
			require.NoError(t, err)
			assert.Equal(t, want, shouldSkipSecrets(cmd))
			assert.False(t, shouldSkipConfig(cmd), "config should still be loaded")
		})
	}
}

func TestGetCalledCommand_MasksSensitiveFlags(t *testing.T) {
	testcases := []struct {
		name          string
		args          string
		wantName      string
		wantFormatted string
	}{
		{name: "no args", args: "", wantName: "porter", wantFormatted: "porter"},
		{name: "no sensitive flags", args: "install --verbosity debug -p myset", wantName: "porter install",
			wantFormatted: "porter install --verbosity debug -p myset"},
		{name: "param", args: "install --param password=topsecret", wantName: "porter install",
			wantFormatted: "porter install --param password=*******"},
		{name: "param with equals", args: "install --param=password=topsecret", wantName: "porter install",
			wantFormatted: "porter install --param=password=*******"},
		{name: "value contains equals", args: "install --param password=top=secret", wantName: "porter install",
			wantFormatted: "porter install --param password=*******"},
		{name: "param without name", args: "install --param topsecret", wantName: "porter install",
			wantFormatted: "porter install --param *******"},
		{name: "multiple params", args: "install mybuns --param name=me -p myset --param password=topsecret --force", wantName: "porter install",
			wantFormatted: "porter install mybuns --param name=******* -p myset --param password=******* --force"},
		{name: "subcommand", args: "installation upgrade --param password=topsecret", wantName: "porter installations upgrade",
			wantFormatted: "porter installation upgrade --param password=*******"},
		{name: "build secret", args: "build --secret id=mysecret,src=/tmp/secret --build-arg A=b --custom c=d", wantName: "porter build",
			wantFormatted: "porter build --secret id=******* --build-arg A=b --custom c=d"},
		{name: "unknown command", args: "instal --param password=topsecret", wantName: "porter",
			wantFormatted: "porter instal --param password=*******"},
		{name: "unknown subcommand", args: "installation instal --param password=topsecret", wantName: "porter installations",
			wantFormatted: "porter installation instal --param password=*******"},
		{name: "positional args", args: "install --param password=topsecret -- --param name=me", wantName: "porter install",
			wantFormatted: "porter install --param password=******* -- --param name=me"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, gotName, gotFormatted := getCalledCommand(buildRootCommand(), strings.Fields(tc.args))
			assert.Equal(t, tc.wantName, gotName)
			assert.Equal(t, tc.wantFormatted, gotFormatted)
		})
	}
}

func TestFormatCommand_Reveal(t *testing.T) {
	reveal := func(flag string, name string) bool {
		return flag == "param" && name == "name"
	}

	args := []string{"install", "--param", "name=me", "--param=name=you", "--param", "password=topsecret", "--param", "name"}
	got := formatCommand(buildRootCommand(), args, reveal)
	assert.Equal(t, "porter install --param name=me --param=name=you --param password=******* --param *******", got)
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
		err := cmd.Execute()
		require.Error(t, err)

		assert.False(t, p.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("flag set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--experimental", experimental.NoopFeature})
		err := cmd.Execute()
		require.Error(t, err)

		assert.True(t, p.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("env set", func(t *testing.T) {
		os.Setenv(expEnvVar, experimental.NoopFeature)
		defer os.Unsetenv(expEnvVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		err := cmd.Execute()
		require.Error(t, err)

		assert.True(t, p.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`experimental: [no-op]`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		err := cmd.Execute()
		require.Error(t, err)

		assert.True(t, p.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("flag set, cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`experimental: []`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--experimental", "no-op"})
		err := cmd.Execute()
		require.Error(t, err)

		assert.True(t, p.IsFeatureEnabled(experimental.FlagNoopFeature))
	})

	t.Run("flag set, env set", func(t *testing.T) {
		os.Setenv(expEnvVar, "")
		defer os.Unsetenv(expEnvVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--experimental", "no-op"})
		err := cmd.Execute()
		require.Error(t, err)

		assert.True(t, p.IsFeatureEnabled(experimental.FlagNoopFeature))
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
		err := cmd.Execute()
		require.Error(t, err)

		assert.False(t, p.IsFeatureEnabled(experimental.FlagNoopFeature))
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
		err := cmd.Execute()
		require.Error(t, err)

		assert.Equal(t, config.LogLevelInfo, p.GetVerbosity())
	})

	t.Run("flag set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--verbosity=debug"})
		err := cmd.Execute()
		require.Error(t, err)

		assert.Equal(t, config.LogLevelDebug, p.GetVerbosity())
	})

	t.Run("env set", func(t *testing.T) {
		os.Setenv(envVar, "error")
		defer os.Unsetenv(envVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		err := cmd.Execute()
		require.Error(t, err)

		assert.Equal(t, config.LogLevelError, p.GetVerbosity())
	})

	t.Run("cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`verbosity: warning`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install"})
		err := cmd.Execute()
		require.Error(t, err)

		assert.Equal(t, config.LogLevelWarn, p.GetVerbosity())
	})

	t.Run("flag set, cfg set", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()

		cfg := []byte(`verbosity: debug`)
		require.NoError(t, p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", cfg, pkg.FileModeWritable))
		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--verbosity", "warn"})
		err := cmd.Execute()
		require.Error(t, err)

		assert.Equal(t, config.LogLevelWarn, p.GetVerbosity())
	})

	t.Run("flag set, env set", func(t *testing.T) {
		os.Setenv(envVar, "warn")
		defer os.Unsetenv(envVar)

		p := porter.NewTestPorter(t)
		defer p.Close()

		cmd := buildRootCommandFrom(p.Porter)
		cmd.SetArgs([]string{"install", "--verbosity=debug"})
		err := cmd.Execute()
		require.Error(t, err)

		assert.Equal(t, config.LogLevelDebug, p.GetVerbosity())
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
		err := cmd.Execute()
		require.Error(t, err)

		assert.Equal(t, config.LogLevelWarn, p.GetVerbosity())
	})
}

func TestPrintError(t *testing.T) {
	testcases := []struct {
		name      string
		verbosity string
		wantStack bool
	}{
		{name: "debug", verbosity: string(config.LogLevelDebug), wantStack: true},
		{name: "info", verbosity: string(config.LogLevelInfo), wantStack: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := porter.NewTestPorter(t)
			defer p.Close()
			p.Data.Verbosity = tc.verbosity

			_, span := tracing.StartSpan(p.RootContext)
			err := span.Error(errors.New("something failed"))
			span.EndSpan()

			var out bytes.Buffer
			printError(&out, p.Porter, err)

			if !tc.wantStack {
				assert.Equal(t, "something failed\n", out.String())
				return
			}
			gotLines := strings.Split(out.String(), "\n")
			require.Greater(t, len(gotLines), 2)
			assert.Equal(t, "something failed", gotLines[0])
			assert.Equal(t, "Stack trace:", gotLines[1])
			assert.Regexp(t, `^  at cmd/porter\.TestPrintError\.func1 \(cmd/porter/main_test\.go:\d+\)$`, gotLines[2])
		})
	}

	t.Run("error without a stack trace", func(t *testing.T) {
		p := porter.NewTestPorter(t)
		defer p.Close()
		p.Data.Verbosity = string(config.LogLevelDebug)

		var out bytes.Buffer
		printError(&out, p.Porter, errors.New("invalid flag"))
		assert.Equal(t, "invalid flag\n", out.String())
	})
}

func TestApplyEarlyVerbosity(t *testing.T) {
	testcases := []struct {
		name string
		args []string
		env  string
		want config.LogLevel
	}{
		{name: "flag with equals", args: []string{"install", "--verbosity=debug"}, want: config.LogLevelDebug},
		{name: "flag with space", args: []string{"install", "--verbosity", "debug", "mybun"}, want: config.LogLevelDebug},
		{name: "other flags", args: []string{"install", "-r", "example/bun:v1", "--param", "a=b", "--force", "--verbosity=debug"}, want: config.LogLevelDebug},
		{name: "env set", args: []string{"install"}, env: "debug", want: config.LogLevelDebug},
		{name: "flag set, env set", args: []string{"install", "--verbosity=debug"}, env: "error", want: config.LogLevelDebug},
		{name: "nothing set", args: []string{"install", "-r", "example/bun:v1"}, want: config.LogLevelWarn},
		{name: "no args", args: nil, want: config.LogLevelWarn},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := porter.NewTestPorter(t)
			defer p.Close()
			// simulate verbosity loaded from the config file
			p.Data.Verbosity = string(config.LogLevelWarn)
			if tc.env != "" {
				p.Setenv(config.EnvPorterVerbosity, tc.env)
			}

			applyEarlyVerbosity(p.Porter, tc.args)
			assert.Equal(t, tc.want, p.GetVerbosity())
		})
	}
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

func TestMarkFlagSensitive(t *testing.T) {
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String("param", "", "")

	markFlagSensitive(f, "param")
	assert.Contains(t, f.Lookup("param").Annotations, sensitiveFlag)

	assert.Panics(t, func() { markFlagSensitive(f, "missing") }, "expected a flag that isn't defined to fail fast")
}
