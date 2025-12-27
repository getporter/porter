package config

import (
	"context"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_GetConfigPath_ExistingToml(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	// Create a config.toml file
	configPath := filepath.Join(c.porterHome, "config.toml")
	err := c.FileSystem.WriteFile(configPath, []byte("verbosity = \"debug\""), pkg.FileModeWritable)
	require.NoError(t, err)

	path, err := c.GetConfigPath()
	require.NoError(t, err)
	assert.Equal(t, configPath, path)
}

func TestConfig_GetConfigPath_ExistingYaml(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	// Create a config.yaml file
	configPath := filepath.Join(c.porterHome, "config.yaml")
	err := c.FileSystem.WriteFile(configPath, []byte("verbosity: debug"), pkg.FileModeWritable)
	require.NoError(t, err)

	path, err := c.GetConfigPath()
	require.NoError(t, err)
	assert.Equal(t, configPath, path)
}

func TestConfig_GetConfigPath_ExistingYml(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	// Create a config.yml file
	configPath := filepath.Join(c.porterHome, "config.yml")
	err := c.FileSystem.WriteFile(configPath, []byte("verbosity: debug"), pkg.FileModeWritable)
	require.NoError(t, err)

	path, err := c.GetConfigPath()
	require.NoError(t, err)
	assert.Equal(t, configPath, path)
}

func TestConfig_GetConfigPath_ExistingJson(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	// Create a config.json file
	configPath := filepath.Join(c.porterHome, "config.json")
	err := c.FileSystem.WriteFile(configPath, []byte(`{"verbosity":"debug"}`), pkg.FileModeWritable)
	require.NoError(t, err)

	path, err := c.GetConfigPath()
	require.NoError(t, err)
	assert.Equal(t, configPath, path)
}

func TestConfig_GetConfigPath_NoExistingConfig(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	path, err := c.GetConfigPath()
	require.NoError(t, err)
	// Should default to config.toml
	assert.Equal(t, filepath.Join(c.porterHome, "config.toml"), path)
}

func TestConfig_GetConfigPath_PriorityOrder(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	// Create multiple config files
	err := c.FileSystem.WriteFile(filepath.Join(c.porterHome, "config.json"), []byte(`{}`), pkg.FileModeWritable)
	require.NoError(t, err)
	err = c.FileSystem.WriteFile(filepath.Join(c.porterHome, "config.yaml"), []byte(""), pkg.FileModeWritable)
	require.NoError(t, err)

	path, err := c.GetConfigPath()
	require.NoError(t, err)
	// Should return config.toml path first if it exists
	// Since we didn't create config.toml, should return config.yaml (second in priority)
	assert.Equal(t, filepath.Join(c.porterHome, "config.yaml"), path)
}

func TestDetectConfigFormat_Toml(t *testing.T) {
	format := DetectConfigFormat("/path/to/config.toml")
	assert.Equal(t, "toml", format)
}

func TestDetectConfigFormat_Yaml(t *testing.T) {
	format := DetectConfigFormat("/path/to/config.yaml")
	assert.Equal(t, encoding.Yaml, format)
}

func TestDetectConfigFormat_Yml(t *testing.T) {
	format := DetectConfigFormat("/path/to/config.yml")
	assert.Equal(t, encoding.Yaml, format)
}

func TestDetectConfigFormat_Json(t *testing.T) {
	format := DetectConfigFormat("/path/to/config.json")
	assert.Equal(t, "json", format)
}

func TestConfig_SaveConfig_Toml(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	ctx := context.Background()
	c.Data.Verbosity = "debug"
	c.Data.Namespace = "test-namespace"

	configPath := filepath.Join(c.porterHome, "config.toml")
	err := c.SaveConfig(ctx, configPath)
	require.NoError(t, err)

	// Verify file was created
	exists, err := c.FileSystem.Exists(configPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify content
	content, err := c.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `verbosity = "debug"`)
	assert.Contains(t, string(content), `namespace = "test-namespace"`)
}

func TestConfig_SaveConfig_Yaml(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	ctx := context.Background()
	c.Data.Verbosity = "info"
	c.Data.Namespace = "prod"

	configPath := filepath.Join(c.porterHome, "config.yaml")
	err := c.SaveConfig(ctx, configPath)
	require.NoError(t, err)

	// Verify file was created
	exists, err := c.FileSystem.Exists(configPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify content
	content, err := c.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "verbosity: info")
	assert.Contains(t, string(content), "namespace: prod")
}

func TestConfig_SaveConfig_Json(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	ctx := context.Background()
	c.Data.Verbosity = "warn"

	configPath := filepath.Join(c.porterHome, "config.json")
	err := c.SaveConfig(ctx, configPath)
	require.NoError(t, err)

	// Verify file was created
	exists, err := c.FileSystem.Exists(configPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify content
	content, err := c.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `"verbosity": "warn"`)
}

func TestConfig_SaveConfig_CreatesParentDirectory(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	ctx := context.Background()
	c.Data.Verbosity = "debug"

	// Use a path with a non-existent subdirectory
	configPath := filepath.Join(c.porterHome, "subdir", "config.toml")
	err := c.SaveConfig(ctx, configPath)
	require.NoError(t, err)

	// Verify file was created
	exists, err := c.FileSystem.Exists(configPath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestConfig_CreateDefaultConfig_Toml(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	ctx := context.Background()
	configPath := filepath.Join(c.porterHome, "config.toml")

	err := c.CreateDefaultConfig(ctx, configPath)
	require.NoError(t, err)

	// Verify file was created
	exists, err := c.FileSystem.Exists(configPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify it has default values
	defaults := DefaultDataStore()
	assert.Equal(t, defaults.BuildDriver, c.Data.BuildDriver)
	assert.Equal(t, defaults.RuntimeDriver, c.Data.RuntimeDriver)
	assert.Equal(t, defaults.DefaultStoragePlugin, c.Data.DefaultStoragePlugin)
	assert.Equal(t, defaults.DefaultSecretsPlugin, c.Data.DefaultSecretsPlugin)
	assert.Equal(t, defaults.Verbosity, c.Data.Verbosity)

	// Verify file content
	content, err := c.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `build-driver = "buildkit"`)
	assert.Contains(t, string(content), `runtime-driver = "docker"`)
}

func TestConfig_CreateDefaultConfig_Yaml(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	ctx := context.Background()
	configPath := filepath.Join(c.porterHome, "config.yaml")

	err := c.CreateDefaultConfig(ctx, configPath)
	require.NoError(t, err)

	// Verify file was created
	exists, err := c.FileSystem.Exists(configPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify file content
	content, err := c.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "build-driver: buildkit")
	assert.Contains(t, string(content), "runtime-driver: docker")
}

func TestConfig_CreateDefaultConfig_Json(t *testing.T) {
	c := NewTestConfig(t)
	defer c.Close()

	ctx := context.Background()
	configPath := filepath.Join(c.porterHome, "config.json")

	err := c.CreateDefaultConfig(ctx, configPath)
	require.NoError(t, err)

	// Verify file was created
	exists, err := c.FileSystem.Exists(configPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify file content
	content, err := c.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `"build-driver": "buildkit"`)
	assert.Contains(t, string(content), `"runtime-driver": "docker"`)
}
