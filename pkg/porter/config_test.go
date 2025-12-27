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

func TestConfigEditOptions_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid with no args",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "invalid with args",
			args:        []string{"somefile"},
			expectError: true,
			errorMsg:    "does not accept arguments",
		},
		{
			name:        "invalid with multiple args",
			args:        []string{"file1", "file2"},
			expectError: true,
			errorMsg:    "does not accept arguments",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := ConfigEditOptions{}

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

func TestPorter_EditConfig_CreatesDefaultWhenMissing(t *testing.T) {
	// This test verifies that EditConfig creates a default config file
	// when none exists. We skip the actual editor interaction.
	p := NewTestPorter(t)
	defer p.Close()

	// Get config path
	configPath, err := p.Config.GetConfigPath()
	require.NoError(t, err)

	// Verify no config exists
	exists, err := p.Config.FileSystem.Exists(configPath)
	require.NoError(t, err)
	require.False(t, exists)

	// Note: We can't easily test the full EditConfig flow without mocking
	// the editor interaction. The editor.Run() call would block waiting for
	// user input. This is better tested via integration tests or manual testing.
	// Here we just verify the preconditions and that CreateDefaultConfig works.

	// Create default config (this is what EditConfig would do)
	err = p.Config.CreateDefaultConfig(context.Background(), configPath)
	require.NoError(t, err)

	// Verify config was created
	exists, err = p.Config.FileSystem.Exists(configPath)
	require.NoError(t, err)
	require.True(t, exists)

	// Verify it contains valid toml
	contents, err := p.Config.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	require.NotEmpty(t, contents)

	// Verify we can parse it
	var data config.Data
	err = encoding.UnmarshalFile(p.Config.FileSystem, configPath, &data)
	require.NoError(t, err)
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
			switch tc.expected {
			case "json":
				require.True(t, strings.HasPrefix(output, "{"))
			case "yaml":
				// YAML typically has key: value format
				require.Contains(t, output, "verbosity: debug")
			case "toml":
				// TOML typically has key = "value" format
				require.Contains(t, output, `verbosity = "debug"`)
			}
		})
	}
}

func TestConfigSetOptions_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
		expectedKey string
		expectedVal string
	}{
		{
			name:        "valid with key and value",
			args:        []string{"verbosity", "debug"},
			expectError: false,
			expectedKey: "verbosity",
			expectedVal: "debug",
		},
		{
			name:        "valid with nested key",
			args:        []string{"logs.level", "info"},
			expectError: false,
			expectedKey: "logs.level",
			expectedVal: "info",
		},
		{
			name:        "invalid with no args",
			args:        []string{},
			expectError: true,
			errorMsg:    "requires exactly 2 arguments",
		},
		{
			name:        "invalid with one arg",
			args:        []string{"verbosity"},
			expectError: true,
			errorMsg:    "requires exactly 2 arguments",
		},
		{
			name:        "invalid with three args",
			args:        []string{"verbosity", "debug", "extra"},
			expectError: true,
			errorMsg:    "requires exactly 2 arguments",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := ConfigSetOptions{}

			err := opts.Validate(tc.args)

			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedKey, opts.Key)
				require.Equal(t, tc.expectedVal, opts.Value)
			}
		})
	}
}

func TestPorter_SetConfig_StringField(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	// Ensure no config exists
	configPath, err := p.Config.GetConfigPath()
	require.NoError(t, err)
	exists, err := p.Config.FileSystem.Exists(configPath)
	require.NoError(t, err)
	require.False(t, exists)

	opts := ConfigSetOptions{
		Key:   "verbosity",
		Value: "debug",
	}

	// Set the value
	err = p.SetConfig(context.Background(), opts)
	require.NoError(t, err)

	// Verify output message
	output := p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "Set verbosity = debug")

	// Verify config file was created
	exists, err = p.Config.FileSystem.Exists(configPath)
	require.NoError(t, err)
	require.True(t, exists)

	// Verify the value was set
	var data config.Data
	err = encoding.UnmarshalFile(p.Config.FileSystem, configPath, &data)
	require.NoError(t, err)
	require.Equal(t, "debug", data.Verbosity)
}

func TestPorter_SetConfig_BoolField(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := ConfigSetOptions{
		Key:   "telemetry.enabled",
		Value: "true",
	}

	// Set the value
	err := p.SetConfig(context.Background(), opts)
	require.NoError(t, err)

	// Verify the value was set
	configPath, err := p.Config.GetConfigPath()
	require.NoError(t, err)
	var data config.Data
	err = encoding.UnmarshalFile(p.Config.FileSystem, configPath, &data)
	require.NoError(t, err)
	require.True(t, data.Telemetry.Enabled)
}

func TestPorter_SetConfig_NestedField(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := ConfigSetOptions{
		Key:   "logs.level",
		Value: "info",
	}

	// Set the value
	err := p.SetConfig(context.Background(), opts)
	require.NoError(t, err)

	// Verify the value was set
	configPath, err := p.Config.GetConfigPath()
	require.NoError(t, err)
	var data config.Data
	err = encoding.UnmarshalFile(p.Config.FileSystem, configPath, &data)
	require.NoError(t, err)
	require.Equal(t, config.LogLevel("info"), data.Logs.Level)
}

func TestPorter_SetConfig_InvalidKey(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := ConfigSetOptions{
		Key:   "invalid.key",
		Value: "value",
	}

	// Set the value
	err := p.SetConfig(context.Background(), opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown config field")
}

func TestPorter_SetConfig_InvalidValue(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := ConfigSetOptions{
		Key:   "logs.level",
		Value: "invalid-level",
	}

	// Set the value
	err := p.SetConfig(context.Background(), opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid value")
}

func TestPorter_SetConfig_PreservesFormat(t *testing.T) {
	testcases := []struct {
		name       string
		configFile string
		content    string
		format     string
	}{
		{
			name:       "toml config",
			configFile: "config.toml",
			content:    `verbosity = "info"`,
			format:     "toml",
		},
		{
			name:       "yaml config",
			configFile: "config.yaml",
			content:    "verbosity: info\n",
			format:     "yaml",
		},
		{
			name:       "json config",
			configFile: "config.json",
			content:    `{"verbosity":"info"}`,
			format:     "json",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			// Write existing config file
			home, err := p.Config.GetHomeDir()
			require.NoError(t, err)
			configPath := filepath.Join(home, tc.configFile)
			err = p.Config.FileSystem.WriteFile(configPath, []byte(tc.content), 0644)
			require.NoError(t, err)

			// Reload config
			_, err = p.Config.Load(context.Background(), nil)
			require.NoError(t, err)

			// Set a new value
			opts := ConfigSetOptions{
				Key:   "namespace",
				Value: "myapp",
			}
			err = p.SetConfig(context.Background(), opts)
			require.NoError(t, err)

			// Verify format is preserved
			detectedFormat := config.DetectConfigFormat(configPath)
			require.Equal(t, tc.format, detectedFormat)

			// Verify the value was set
			var data config.Data
			err = encoding.UnmarshalFile(p.Config.FileSystem, configPath, &data)
			require.NoError(t, err)
			require.Equal(t, "myapp", data.Namespace)
		})
	}
}

func TestPorter_SetConfig_MultipleValues(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	// Set multiple values
	values := []struct {
		key   string
		value string
	}{
		{"verbosity", "debug"},
		{"namespace", "test"},
		{"logs.level", "info"},
		{"telemetry.enabled", "true"},
	}

	for _, v := range values {
		opts := ConfigSetOptions{
			Key:   v.key,
			Value: v.value,
		}
		err := p.SetConfig(context.Background(), opts)
		require.NoError(t, err)
	}

	// Verify all values were set
	configPath, err := p.Config.GetConfigPath()
	require.NoError(t, err)
	var data config.Data
	err = encoding.UnmarshalFile(p.Config.FileSystem, configPath, &data)
	require.NoError(t, err)

	require.Equal(t, "debug", data.Verbosity)
	require.Equal(t, "test", data.Namespace)
	require.Equal(t, config.LogLevel("info"), data.Logs.Level)
	require.True(t, data.Telemetry.Enabled)
}
