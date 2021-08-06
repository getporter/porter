package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromConfigFile(t *testing.T) {
	c := NewTestConfig(t)
	c.SetHomeDir("/root/.porter")

	c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

	c.DataLoader = LoadFromEnvironment()
	err := c.LoadData()
	require.NoError(t, err, "dataloader failed")
	assert.True(t, c.Debug, "config.Debug was not set correctly")
}

func TestData_Marshal(t *testing.T) {
	c := NewTestConfig(t)
	c.SetHomeDir("/root/.porter")

	c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

	c.DataLoader = LoadFromEnvironment()
	err := c.LoadData()
	require.NoError(t, err, "LoadData failed")

	// Check Storage Attributes
	assert.Equal(t, "dev", c.Data.DefaultStorage, "DefaultStorage was not loaded properly")
	assert.Equal(t, "azure.blob", c.Data.DefaultStoragePlugin, "DefaultStoragePlugin was not loaded properly")

	require.Len(t, c.Data.StoragePlugins, 1, "StoragePlugins was not loaded properly")
	devStore := c.Data.StoragePlugins[0]
	assert.Equal(t, "dev", devStore.Name, "StoragePlugins.Name was not loaded properly")
	assert.Equal(t, "azure.blob", devStore.PluginSubKey, "StoragePlugins.PluginSubKey was not loaded properly")
	assert.Equal(t, map[string]interface{}{"env": "DEV_AZURE_STORAGE_CONNECTION_STRING"}, devStore.Config, "StoragePlugins.Config was not loaded properly")

	// Check Secret Attributes
	assert.Equal(t, "red-team", c.Data.DefaultSecrets, "DefaultSecrets was not loaded properly")
	assert.Equal(t, "azure.keyvault", c.Data.DefaultSecretsPlugin, "DefaultSecretsPlugin was not loaded properly")

	require.Len(t, c.Data.SecretsPlugin, 1, "SecretsPlugins was not loaded properly")
	teamSource := c.Data.SecretsPlugin[0]
	assert.Equal(t, "red-team", teamSource.Name, "SecretsPlugins.Name was not loaded properly")
	assert.Equal(t, "azure.keyvault", teamSource.PluginSubKey, "SecretsPlugins.PluginSubKey was not loaded properly")
	assert.Equal(t, map[string]interface{}{"vault": "teamsekrets"}, teamSource.Config, "SecretsPlugins.Config was not loaded properly")
}
