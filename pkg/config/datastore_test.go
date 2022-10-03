package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestData_GetDefaultStoragePlugin(t *testing.T) {
	c := New()
	assert.Equal(t, "mongodb-docker", c.Data.DefaultStoragePlugin, "Built-in mongodb-docker plugin should be used when config is missing")
}

func TestData_StorageAccessors(t *testing.T) {
	c := Config{
		Data: Data{
			DefaultStoragePlugin: "blorpy",
			DefaultStorage:       "dev",
			StoragePlugins: []StoragePlugin{
				{PluginConfig{
					Name:         "dev",
					PluginSubKey: "hashicorp.vault",
				}},
			},
		},
	}

	assert.Equal(t, "blorpy", c.Data.DefaultStoragePlugin, "GetDefaultStoragePlugin returned the wrong value")
	assert.Equal(t, "dev", c.Data.DefaultStorage, "GetDefaultStorage returned the wrong value")

	store, err := c.GetStorage("dev")
	require.NoError(t, err, "GetStorage failed")
	require.NotNil(t, store, "GetStorage returned a nil StoragePlugin")
	assert.Equal(t, "dev", store.Name, "StoragePlugin.Name returned the wrong value")
	assert.Equal(t, "hashicorp.vault", store.PluginSubKey, "StoragePlugin.PluginSubKey returned the wrong value")
}

func TestData_GetDefaultSecretsPlugin(t *testing.T) {
	c := New()
	assert.Equal(t, "host", c.Data.DefaultSecretsPlugin, "Built-in host plugin should be used when config is missing")
}

func TestData_SecretAccessors(t *testing.T) {
	c := Config{
		Data: Data{
			DefaultSecretsPlugin: "topsekret",
			DefaultSecrets:       "prod",
			SecretsPlugin: []SecretsPlugin{
				{PluginConfig{
					Name:         "prod",
					PluginSubKey: "azure.keyvault",
				}},
			},
		},
	}

	assert.Equal(t, "topsekret", c.Data.DefaultSecretsPlugin, "GetDefaultSecretsPlugin returned the wrong value")
	assert.Equal(t, "prod", c.Data.DefaultSecrets, "GetDefaultSecretsPlugin returned the wrong value")

	source, err := c.GetSecretsPlugin("prod")
	require.NoError(t, err, "GetSecretsPlugin failed")
	require.NotNil(t, source, "GetSecretsPlugin returned a nil SecretsPlugin")
	assert.Equal(t, "prod", source.Name, "SecretsPlugin.Name returned the wrong value")
	assert.Equal(t, "azure.keyvault", source.PluginSubKey, "SecretsPlugin.PluginSubKey returned the wrong value")
}
