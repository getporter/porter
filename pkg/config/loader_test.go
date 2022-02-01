package config

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_OverrideWithEnvironmentVariable(t *testing.T) {
	// Do not run in parallel, it sets environment variables
	os.Setenv("PORTER_DEFAULT_STORAGE", "")
	defer os.Unsetenv("PORTER_DEFAULT_STORAGE")

	c := NewTestConfig(t)
	c.SetHomeDir("/root/.porter")

	c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

	c.DataLoader = LoadFromEnvironment()
	err := c.Load(context.Background(), nil)

	require.NoError(t, err, "dataloader failed")
	assert.True(t, c.Debug, "config.Debug was not set correctly")
	assert.Empty(t, c.Data.DefaultStorage, "The config file value should be overridden by an empty env var")
}

func TestData_Marshal(t *testing.T) {
	// Do not run in parallel, it sets environment variables
	os.Setenv("VAULT_TOKEN", "topsecret-token")
	defer os.Unsetenv("VAULT_TOKEN")

	c := NewTestConfig(t)
	c.SetHomeDir("/root/.porter")

	c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

	c.DataLoader = LoadFromEnvironment()
	resolveTestSecrets := func(secretKey string) (string, error) {
		return "topsecret-connectionstring", nil
	}
	err := c.Load(context.Background(), resolveTestSecrets)
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
