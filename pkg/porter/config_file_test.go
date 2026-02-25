package porter

import (
	"context"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigFilePath_NoConfigExists(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	path, exists, err := p.GetConfigFilePath()
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Contains(t, path, "config.yaml")
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

func TestConfigContextList(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.yaml")
	configContent := `schemaVersion: "` + config.ConfigSchemaVersion + `"
current-context: prod
contexts:
  - name: default
    config: {}
  - name: prod
    config: {}
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))

	err := p.ConfigContextList(context.Background())
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "* prod")
	assert.Contains(t, output, "  default")
}

func TestConfigContextList_FlagOverridesCurrentContext(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.yaml")
	configContent := `schemaVersion: "` + config.ConfigSchemaVersion + `"
current-context: prod
contexts:
  - name: default
    config: {}
  - name: prod
    config: {}
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))
	p.Config.ContextName = "default"

	err := p.ConfigContextList(context.Background())
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "* default")
	assert.Contains(t, output, "  prod")
}

func TestConfigContextList_NoFile(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	err := p.ConfigContextList(context.Background())
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "No configuration file found")
}

func TestConfigContextList_LegacyFile(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.toml")
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(`verbosity = "debug"`), 0600))

	err := p.ConfigContextList(context.Background())
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "legacy flat format")
}

func TestConfigContextList_EmptyContexts(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.yaml")
	configContent := `schemaVersion: "` + config.ConfigSchemaVersion + `"
current-context: default
contexts: []
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))

	err := p.ConfigContextList(context.Background())
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "No contexts defined")
}

func TestConfigContextUse(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.yaml")
	configContent := `schemaVersion: "` + config.ConfigSchemaVersion + `"
current-context: default
contexts:
  - name: default
    config: {}
  - name: prod
    config: {}
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))

	err := p.ConfigContextUse(context.Background(), "prod")
	require.NoError(t, err)

	updated, err := p.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(updated), "current-context: prod")
	assert.NotContains(t, string(updated), "current-context: default")
}

func TestConfigContextUse_NoFile(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	err := p.ConfigContextUse(context.Background(), "prod")
	require.ErrorContains(t, err, "no config file found")
}

func TestConfigMigrate(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.yaml")
	legacy := `namespace: dev
verbosity: debug
storage:
  - name: testdb
    plugin: mongodb
    config:
      url: mongodb://localhost:27017/${env.PORTER_TEST_DB_NAME}?connect=direct
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(legacy), 0600))

	err := p.ConfigMigrate(context.Background())
	require.NoError(t, err)

	result, err := p.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	content := string(result)

	assert.Contains(t, content, `schemaVersion: "2.0.0"`)
	assert.Contains(t, content, "current-context: default")
	assert.Contains(t, content, "- name: default")
	assert.Contains(t, content, "      namespace: dev")           // indented 6 spaces
	assert.Contains(t, content, "${env.PORTER_TEST_DB_NAME}")    // template var preserved
	assert.NotContains(t, content, "\nnamespace: dev")           // top-level key must be gone

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "Migrated")
}

func TestConfigMigrate_AlreadyMigrated(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.yaml")
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(`schemaVersion: "2.0.0"
current-context: default
contexts:
  - name: default
    config: {}
`), 0600))

	err := p.ConfigMigrate(context.Background())
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "already using the multi-context format")
}

func TestConfigMigrate_NoFile(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	err := p.ConfigMigrate(context.Background())
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "No configuration file found")
}

func TestConfigMigrate_NonYAML(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.toml")
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(`namespace = "dev"`), 0600))

	err := p.ConfigMigrate(context.Background())
	require.ErrorContains(t, err, "toml")
	require.ErrorContains(t, err, "manually")
}

func TestConfigContextUse_LegacyFile(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.toml")
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(`verbosity = "debug"`), 0600))

	err := p.ConfigContextUse(context.Background(), "prod")
	require.ErrorContains(t, err, "not a versioned multi-context file")
}

func TestConfigContextUse_NonexistentContext(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.yaml")
	configContent := `schemaVersion: "` + config.ConfigSchemaVersion + `"
current-context: default
contexts:
  - name: default
    config: {}
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))

	err := p.ConfigContextUse(context.Background(), "nonexistent")
	require.ErrorContains(t, err, `"nonexistent" not found`)
}

func TestConfigContextUse_InvalidName(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.yaml")
	configContent := `schemaVersion: "` + config.ConfigSchemaVersion + `"
current-context: default
contexts:
  - name: default
    config: {}
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))

	for _, name := range []string{
		"bad\nname",  // newline
		"bad name",  // space
		"bad#name",  // hash (comment in YAML)
		"bad\"name", // double-quote (breaks TOML/JSON)
		"#leading",  // must start with letter/digit
	} {
		err := p.ConfigContextUse(context.Background(), name)
		require.ErrorContains(t, err, "invalid context name", "expected error for name %q", name)
	}
}

func TestConfigContextUse_TOML(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.toml")
	configContent := `schemaVersion = "` + config.ConfigSchemaVersion + `"
current-context = "default"

[[contexts]]
name = "default"

[[contexts]]
name = "prod"
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))

	err := p.ConfigContextUse(context.Background(), "prod")
	require.NoError(t, err)

	updated, err := p.FileSystem.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(updated), `current-context = "prod"`)
	assert.NotContains(t, string(updated), `current-context = "default"`)
}

// TestConfigContextList_TOMLMultiContext verifies that ConfigContextList works
// with a manually-created TOML multi-context file. Auto-migration is YAML-only,
// but TOML (and other viper-supported formats) remain fully supported for reading.
func TestConfigContextList_TOMLMultiContext(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	home, _ := p.GetHomeDir()
	configPath := filepath.Join(home, "config.toml")
	// TOML equivalent of the multi-context format, written manually.
	configContent := `schemaVersion = "` + config.ConfigSchemaVersion + `"
current-context = "prod"

[[contexts]]
name = "default"

[[contexts]]
name = "prod"
`
	require.NoError(t, p.FileSystem.WriteFile(configPath, []byte(configContent), 0600))

	err := p.ConfigContextList(context.Background())
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "* prod")
	assert.Contains(t, output, "  default")
}
