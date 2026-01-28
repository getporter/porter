package porter

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigFilePath_NoConfigExists(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	path, exists, err := p.GetConfigFilePath()
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Contains(t, path, "config.toml")
}

func TestGetConfigFilePath_TomlExists(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.toml")
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte("verbosity = \"debug\""), 0600))

	path, exists, err := p.GetConfigFilePath()
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, configPath, path)
}

func TestGetConfigFilePath_YamlExists(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.yaml")
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte("verbosity: debug"), 0600))

	path, exists, err := p.GetConfigFilePath()
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, configPath, path)
}

func TestGetConfigFilePath_TomlTakesPrecedence(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	tomlPath := filepath.Join(home, "config.toml")
	yamlPath := filepath.Join(home, "config.yaml")
	require.NoError(t, p.FileSystem.WriteFile(tomlPath, []byte("verbosity = \"debug\""), 0600))
	require.NoError(t, p.FileSystem.WriteFile(yamlPath, []byte("verbosity: debug"), 0600))

	path, exists, err := p.GetConfigFilePath()
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, tomlPath, path)
}

func TestConfigShow_NoConfig(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	err := p.ConfigShow(context.Background(), ConfigShowOptions{})
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "No configuration file found")
	assert.Contains(t, output, "porter config edit")
}

func TestConfigShow_WithConfig(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.toml")
	configContent := `verbosity = "debug"
namespace = "test"
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))

	err := p.ConfigShow(context.Background(), ConfigShowOptions{})
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, `verbosity = "debug"`)
	assert.Contains(t, output, `namespace = "test"`)
}

func TestConfigShow_PreservesTemplateVars(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.toml")
	configContent := `verbosity = "${env.PORTER_VERBOSITY}"
default-secrets-plugin = "${secret.PLUGIN_NAME}"
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))

	err := p.ConfigShow(context.Background(), ConfigShowOptions{})
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "${env.PORTER_VERBOSITY}")
	assert.Contains(t, output, "${secret.PLUGIN_NAME}")
}
