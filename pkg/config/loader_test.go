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

	require.Len(t, c.Data.CrudStores, 1, "CrudStores was not loaded properly")
	devStore := c.Data.CrudStores[0]
	assert.Equal(t, "dev", devStore.Name, "CrudStores.Name was not loaded properly")
	assert.Equal(t, "azure.blob", devStore.PluginSubKey, "CrudStores.PluginSubKey was not loaded properly")
	assert.Equal(t, map[string]interface{}{"env": "DEV_AZURE_STORAGE_CONNECTION_STRING"}, devStore.Config, "CrudStores.Config was not loaded properly")

	// Check Secret Attributes
	assert.Equal(t, "red-team", c.Data.DefaultSecrets, "DefaultSecrets was not loaded properly")
	assert.Equal(t, "azure.keyvault", c.Data.DefaultSecretsPlugin, "DefaultSecretsPlugin was not loaded properly")

	require.Len(t, c.Data.SecretSources, 1, "SecretSources was not loaded properly")
	teamSource := c.Data.SecretSources[0]
	assert.Equal(t, "red-team", teamSource.Name, "SecretSources.Name was not loaded properly")
	assert.Equal(t, "azure.keyvault", teamSource.PluginSubKey, "SecretSources.PluginSubKey was not loaded properly")
	assert.Equal(t, map[string]interface{}{"vault": "teamsekrets"}, teamSource.Config, "SecretSources.Config was not loaded properly")
}
