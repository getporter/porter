package config

import (
	"context"
	"os"
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_OverrideWithEnvironmentVariable(t *testing.T) {
	// Do not run in parallel, it sets environment variables
	os.Setenv("PORTER_DEFAULT_STORAGE", "")
	defer os.Unsetenv("PORTER_DEFAULT_STORAGE")

	c := NewTestConfig(t)
	c.SetHomeDir("/home/myuser/.porter")

	c.TestContext.AddTestFile("testdata/config.toml", "/home/myuser/.porter/config.toml")

	c.DataLoader = LoadFromEnvironment()
	_, err := c.Load(context.Background(), nil)

	require.NoError(t, err, "dataloader failed")
	assert.Equal(t, "warn", c.Data.Verbosity, "config.Verbosity was not set correctly")
	assert.Empty(t, c.Data.DefaultStorage, "The config file value should be overridden by an empty env var")
}

func TestData_Marshal(t *testing.T) {
	// Do not run in parallel, it sets environment variables
	os.Setenv("VAULT_TOKEN", "topsecret-token")
	defer os.Unsetenv("VAULT_TOKEN")

	c := NewTestConfig(t)
	c.SetHomeDir("/home/myuser/.porter")

	c.TestContext.AddTestFile("testdata/config.toml", "/home/myuser/.porter/config.toml")

	c.DataLoader = LoadFromEnvironment()
	resolveTestSecrets := func(ctx context.Context, secretKey string) (string, error) {
		return "topsecret-connectionstring", nil
	}
	_, err := c.Load(context.Background(), resolveTestSecrets)
	require.NoError(t, err, "Load failed")

	// Check Storage Attributes
	assert.Equal(t, "dev", c.Data.DefaultStorage, "DefaultStorage was not loaded properly")
	assert.Equal(t, "mongodb-docker", c.Data.DefaultStoragePlugin, "DefaultStoragePlugin was not loaded properly")

	require.Len(t, c.Data.StoragePlugins, 1, "StoragePlugins was not loaded properly")
	devStore := c.Data.StoragePlugins[0]
	assert.Equal(t, "dev", devStore.Name, "StoragePlugins.Name was not loaded properly")
	assert.Equal(t, "mongodb", devStore.PluginSubKey, "StoragePlugins.PluginSubKey was not loaded properly")
	assert.Equal(t, map[string]interface{}{"url": "topsecret-connectionstring"}, devStore.Config, "StoragePlugins.Config was not loaded properly")

	// Check Secret Attributes
	assert.Equal(t, "red-team", c.Data.DefaultSecrets, "DefaultSecrets was not loaded properly")
	assert.Equal(t, "azure.keyvault", c.Data.DefaultSecretsPlugin, "DefaultSecretsPlugin was not loaded properly")

	require.Len(t, c.Data.SecretsPlugin, 1, "SecretsPlugins was not loaded properly")
	teamSource := c.Data.SecretsPlugin[0]
	assert.Equal(t, "red-team", teamSource.Name, "SecretsPlugins.Name was not loaded properly")
	assert.Equal(t, "azure.keyvault", teamSource.PluginSubKey, "SecretsPlugins.PluginSubKey was not loaded properly")
	assert.Equal(t, map[string]interface{}{"token": "topsecret-token", "vault": "teamsekrets"}, teamSource.Config, "SecretsPlugins.Config was not loaded properly")
}

func TestLoadMultiContext_DefaultContext(t *testing.T) {
	t.Parallel()

	c := NewTestConfig(t)
	c.SetHomeDir("/home/myuser/.porter")
	c.TestContext.AddTestFile("testdata/config-multi-context.yaml", "/home/myuser/.porter/config.yaml")

	c.DataLoader = LoadFromFilesystem()
	_, err := c.Load(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, "dev", c.Data.Namespace)
	assert.Equal(t, "debug", c.Data.Verbosity)
	assert.Equal(t, "localdb", c.Data.DefaultStorage)
	assert.Equal(t, "filesystem", c.Data.DefaultSecretsPlugin)
	require.Len(t, c.Data.StoragePlugins, 1)
	assert.Equal(t, "localdb", c.Data.StoragePlugins[0].Name)
}

func TestLoadMultiContext_FlagOverridesDefault(t *testing.T) {
	t.Parallel()

	c := NewTestConfig(t)
	c.SetHomeDir("/home/myuser/.porter")
	c.TestContext.AddTestFile("testdata/config-multi-context.yaml", "/home/myuser/.porter/config.yaml")
	c.ContextName = "prod"

	c.DataLoader = LoadFromFilesystem()
	_, err := c.Load(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, "prod", c.Data.Namespace)
	assert.Equal(t, "warn", c.Data.Verbosity)
	assert.Equal(t, "proddb", c.Data.DefaultStorage)
	assert.Equal(t, "azure.keyvault", c.Data.DefaultSecretsPlugin)
	require.Len(t, c.Data.StoragePlugins, 1)
	assert.Equal(t, "proddb", c.Data.StoragePlugins[0].Name)
}

func TestLoadMultiContext_EnvVarSelectsContext(t *testing.T) {
	// Do not run in parallel — sets environment variables
	os.Setenv("PORTER_CONTEXT", "prod")
	defer os.Unsetenv("PORTER_CONTEXT")

	c := NewTestConfig(t)
	c.SetHomeDir("/home/myuser/.porter")
	c.TestContext.AddTestFile("testdata/config-multi-context.yaml", "/home/myuser/.porter/config.yaml")

	c.DataLoader = LoadFromEnvironment()
	_, err := c.Load(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, "prod", c.Data.Namespace)
	assert.Equal(t, "proddb", c.Data.DefaultStorage)
}

func TestLoadMultiContext_ContextNotFound(t *testing.T) {
	t.Parallel()

	c := NewTestConfig(t)
	c.SetHomeDir("/home/myuser/.porter")
	c.TestContext.AddTestFile("testdata/config-multi-context.yaml", "/home/myuser/.porter/config.yaml")
	c.ContextName = "staging"

	c.DataLoader = LoadFromFilesystem()
	_, err := c.Load(context.Background(), nil)
	require.ErrorContains(t, err, `context "staging" not found`)
}

func TestLoadMultiContext_LegacyFlagError(t *testing.T) {
	t.Parallel()

	c := NewTestConfig(t)
	c.SetHomeDir("/home/myuser/.porter")
	c.TestContext.AddTestFile("testdata/config.toml", "/home/myuser/.porter/config.toml")
	c.ContextName = "prod"

	c.DataLoader = LoadFromFilesystem()
	_, err := c.Load(context.Background(), nil)
	require.ErrorContains(t, err, "--context")
}

func TestLoadMultiContext_MissingContextsKey(t *testing.T) {
	t.Parallel()

	c := NewTestConfig(t)
	c.SetHomeDir("/home/myuser/.porter")

	cfg := `schemaVersion: "` + ConfigSchemaVersion + `"
current-context: default
`
	require.NoError(t, c.TestContext.FileSystem.WriteFile(
		"/home/myuser/.porter/config.yaml", []byte(cfg), 0600))

	c.DataLoader = LoadFromFilesystem()
	_, err := c.Load(context.Background(), nil)
	require.ErrorContains(t, err, "missing required 'contexts' key")
}

func TestListTemplateVariables(t *testing.T) {
	eng := liquid.NewEngine()
	tmpl, err := eng.ParseString(`not a variable {{secrets.foo}} more non variable junk{{env.var}}{{env.var}}`)
	require.NoError(t, err)

	vars := listTemplateVariables(tmpl)
	assert.Equal(t, []string{"env.var", "secrets.foo"}, vars)
}
