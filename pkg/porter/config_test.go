package porter

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/printer"
	"github.com/stretchr/testify/require"
)

func TestConfigShowOptions_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		args        []string
		rawFormat   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid with no args, no format",
			args:        []string{},
			rawFormat:   "",
			expectError: false,
		},
		{
			name:        "valid with json format",
			args:        []string{},
			rawFormat:   "json",
			expectError: false,
		},
		{
			name:        "valid with yaml format",
			args:        []string{},
			rawFormat:   "yaml",
			expectError: false,
		},
		{
			name:        "valid with toml format",
			args:        []string{},
			rawFormat:   "toml",
			expectError: false,
		},
		{
			name:        "invalid format",
			args:        []string{},
			rawFormat:   "xml",
			expectError: true,
			errorMsg:    "invalid format",
		},
		{
			name:        "invalid with args",
			args:        []string{"somefile"},
			rawFormat:   "",
			expectError: true,
			errorMsg:    "does not accept arguments",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := ConfigShowOptions{
				PrintOptions: printer.PrintOptions{
					RawFormat: tc.rawFormat,
				},
			}

			err := opts.Validate(tc.args)

			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPorter_ShowConfig_NoConfigExists(t *testing.T) {
	testcases := []struct {
		name   string
		format printer.Format
	}{
		{name: "toml", format: "toml"},
		{name: "json", format: printer.FormatJson},
		{name: "yaml", format: printer.FormatYaml},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			opts := ConfigShowOptions{
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
			}

			// Ensure no config exists
			configPath, err := p.Config.GetConfigPath()
			require.NoError(t, err)
			exists, err := p.Config.FileSystem.Exists(configPath)
			require.NoError(t, err)
			require.False(t, exists)

			// Show config
			err = p.ShowConfig(context.Background(), opts)
			require.NoError(t, err)

			// Verify output contains default values
			output := p.TestConfig.TestContext.GetOutput()
			require.NotEmpty(t, output)

			// Verify it's valid format by unmarshaling
			var data config.Data
			err = encoding.Unmarshal(string(tc.format), []byte(output), &data)
			require.NoError(t, err)

			// Verify it has default values
			defaultData := config.DefaultDataStore()
			require.Equal(t, defaultData.Namespace, data.Namespace)
		})
	}
}

func TestPorter_ShowConfig_WithExistingConfig(t *testing.T) {
	testcases := []struct {
		name       string
		configFile string
		format     string
		content    string
	}{
		{
			name:       "toml config, show as toml",
			configFile: "config.toml",
			format:     "toml",
			content:    `verbosity = "debug"`,
		},
		{
			name:       "yaml config, show as yaml",
			configFile: "config.yaml",
			format:     "yaml",
			content:    "verbosity: debug\n",
		},
		{
			name:       "json config, show as json",
			configFile: "config.json",
			format:     "json",
			content:    `{"verbosity":"debug"}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			// Write a config file
			home, err := p.Config.GetHomeDir()
			require.NoError(t, err)
			configPath := filepath.Join(home, tc.configFile)
			err = p.Config.FileSystem.WriteFile(configPath, []byte(tc.content), 0644)
			require.NoError(t, err)

			// Reload config to pick up the file
			_, err = p.Config.Load(context.Background(), nil)
			require.NoError(t, err)

			opts := ConfigShowOptions{
				PrintOptions: printer.PrintOptions{
					RawFormat: tc.format,
					Format:    printer.Format(tc.format),
				},
			}

			// Show config
			err = p.ShowConfig(context.Background(), opts)
			require.NoError(t, err)

			// Verify output
			output := p.TestConfig.TestContext.GetOutput()
			require.NotEmpty(t, output)

			// Verify it contains the expected verbosity
			require.Contains(t, output, "debug")

			// Verify it's valid format by unmarshaling
			var data config.Data
			err = encoding.Unmarshal(tc.format, []byte(output), &data)
			require.NoError(t, err)
			require.Equal(t, "debug", data.Verbosity)
		})
	}
}

func TestPorter_ShowConfig_FormatConversion(t *testing.T) {
	// Create a toml config and show it as json and yaml
	p := NewTestPorter(t)
	defer p.Close()

	// Write a toml config file
	home, err := p.Config.GetHomeDir()
	require.NoError(t, err)
	configPath := filepath.Join(home, "config.toml")
	err = p.Config.FileSystem.WriteFile(configPath, []byte(`verbosity = "info"`), 0644)
	require.NoError(t, err)

	// Reload config
	_, err = p.Config.Load(context.Background(), nil)
	require.NoError(t, err)

	testcases := []struct {
		name   string
		format string
	}{
		{name: "convert to json", format: "json"},
		{name: "convert to yaml", format: "yaml"},
		{name: "keep as toml", format: "toml"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset output buffer
			p.TestConfig.TestContext.ClearOutputs()

			opts := ConfigShowOptions{
				PrintOptions: printer.PrintOptions{
					RawFormat: tc.format,
					Format:    printer.Format(tc.format),
				},
			}

			// Show config in requested format
			err = p.ShowConfig(context.Background(), opts)
			require.NoError(t, err)

			// Verify output
			output := p.TestConfig.TestContext.GetOutput()
			require.NotEmpty(t, output)

			// Verify it's valid format by unmarshaling
			var data config.Data
			err = encoding.Unmarshal(tc.format, []byte(output), &data)
			require.NoError(t, err)
			require.Equal(t, "info", data.Verbosity)
		})
	}
}

func TestPorter_ShowConfig_PreservesFormat(t *testing.T) {
	// When no explicit format is requested, existing file format should be used
	testcases := []struct {
		name       string
		configFile string
		content    string
		expected   string
	}{
		{
			name:       "toml config",
			configFile: "config.toml",
			content:    `verbosity = "debug"`,
			expected:   "toml",
		},
		{
			name:       "yaml config",
			configFile: "config.yaml",
			content:    "verbosity: debug\n",
			expected:   "yaml",
		},
		{
			name:       "json config",
			configFile: "config.json",
			content:    `{"verbosity":"debug"}`,
			expected:   "json",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			// Write a config file
			home, err := p.Config.GetHomeDir()
			require.NoError(t, err)
			configPath := filepath.Join(home, tc.configFile)
			err = p.Config.FileSystem.WriteFile(configPath, []byte(tc.content), 0644)
			require.NoError(t, err)

			// Reload config
			_, err = p.Config.Load(context.Background(), nil)
			require.NoError(t, err)

			opts := ConfigShowOptions{
				PrintOptions: printer.PrintOptions{
					RawFormat: "", // No explicit format
				},
			}

			// Show config
			err = p.ShowConfig(context.Background(), opts)
			require.NoError(t, err)

			// Verify output is in the original format
			output := strings.TrimSpace(p.TestConfig.TestContext.GetOutput())
			require.NotEmpty(t, output)

			// Simple format detection by checking output structure
			if tc.expected == "json" {
				require.True(t, strings.HasPrefix(output, "{"))
			} else if tc.expected == "yaml" {
				// YAML typically has key: value format
				require.Contains(t, output, "verbosity: debug")
			} else if tc.expected == "toml" {
				// TOML typically has key = "value" format
				require.Contains(t, output, `verbosity = "debug"`)
			}
		})
	}
}
