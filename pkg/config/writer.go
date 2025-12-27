package config

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/tracing"
)

// GetConfigPath returns the path to the Porter config file.
// It checks PORTER_HOME for config files in order: config.toml, config.yaml, config.yml, config.json
// Returns the first found, or defaults to config.toml path if none exist.
// Returns error if PORTER_HOME cannot be determined.
func (c *Config) GetConfigPath() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine PORTER_HOME: %w", err)
	}

	// Check for existing config files in priority order
	configFiles := []string{"config.toml", "config.yaml", "config.yml", "config.json"}
	for _, filename := range configFiles {
		path := filepath.Join(home, filename)
		exists, err := c.FileSystem.Exists(path)
		if err != nil {
			return "", fmt.Errorf("error checking if %s exists: %w", path, err)
		}
		if exists {
			return path, nil
		}
	}

	// Default to config.toml if no config exists
	return filepath.Join(home, "config.toml"), nil
}

// DetectConfigFormat extracts the file extension from the path and returns
// the format string: "toml", "yaml", or "json".
func DetectConfigFormat(path string) string {
	ext := filepath.Ext(path)
	format := strings.TrimPrefix(ext, ".")

	// Normalize yml to yaml
	if format == "yml" {
		return encoding.Yaml
	}

	return format
}

// SaveConfig marshals the config data to the specified path using the format
// determined from the file extension. Creates parent directories if needed.
func (c *Config) SaveConfig(ctx context.Context, path string) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := c.FileSystem.MkdirAll(dir, pkg.FileModeDirectory); err != nil {
		return log.Error(fmt.Errorf("error creating config directory %s: %w", dir, err))
	}

	// Detect format from path extension
	format := DetectConfigFormat(path)

	// Marshal config data
	data, err := encoding.Marshal(format, c.Data)
	if err != nil {
		return log.Error(fmt.Errorf("error marshaling config to %s: %w", format, err))
	}

	// Write to file
	if err := c.FileSystem.WriteFile(path, data, pkg.FileModeWritable); err != nil {
		return log.Error(fmt.Errorf("error writing config file %s: %w", path, err))
	}

	log.Infof("Config saved to %s", path)
	return nil
}

// CreateDefaultConfig creates a new config file with default values at the specified path.
func (c *Config) CreateDefaultConfig(ctx context.Context, path string) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// Set config data to defaults
	c.Data = DefaultDataStore()

	// Use SaveConfig to handle marshaling and writing
	if err := c.SaveConfig(ctx, path); err != nil {
		return log.Error(fmt.Errorf("error creating default config: %w", err))
	}

	log.Infof("Created default config at %s", path)
	return nil
}
