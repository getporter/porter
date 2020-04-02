package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestData_GetDefaultStoragePlugin(t *testing.T) {
	var d Data
	assert.Equal(t, "filesystem", d.GetDefaultStoragePlugin(), "Built-in filesystem plugin should be used when config is missing")
}

func TestData_StorageAccessors(t *testing.T) {
	d := Data{
		DefaultStoragePlugin: "blorpy",
		DefaultStorage:       "dev",
		CrudStores: []CrudStore{
			{PluginConfig{
				Name:         "dev",
				PluginSubKey: "hashicorp.vault",
			}},
		},
	}

	assert.Equal(t, "blorpy", d.GetDefaultStoragePlugin(), "GetDefaultStoragePlugin returned the wrong value")
	assert.Equal(t, "dev", d.GetDefaultStorage(), "GetDefaultStorage returned the wrong value")

	store, err := d.GetStorage("dev")
	require.NoError(t, err, "GetStorage failed")
	require.NotNil(t, store, "GetStorage returned a nil CrudStore")
	assert.Equal(t, "dev", store.Name, "CrudStore.Name returned the wrong value")
	assert.Equal(t, "hashicorp.vault", store.PluginSubKey, "CrudStore.PluginSubKey returned the wrong value")
}

func TestData_GetDefaultSecretsPlugin(t *testing.T) {
	var d Data
	assert.Equal(t, "host", d.GetDefaultSecretsPlugin(), "Built-in host plugin should be used when config is missing")
}

func TestData_SecretAccessors(t *testing.T) {
	d := Data{
		DefaultSecretsPlugin: "topsekret",
		DefaultSecrets:       "prod",
		SecretSources: []SecretSource{
			{PluginConfig{
				Name:         "prod",
				PluginSubKey: "azure.keyvault",
			}},
		},
	}

	assert.Equal(t, "topsekret", d.GetDefaultSecretsPlugin(), "GetDefaultSecretsPlugin returned the wrong value")
	assert.Equal(t, "prod", d.GetDefaultSecretSource(), "GetDefaultSecretSource returned the wrong value")

	source, err := d.GetSecretSource("prod")
	require.NoError(t, err, "GetSecretSource failed")
	require.NotNil(t, source, "GetSecretSource returned a nil SecretSource")
	assert.Equal(t, "prod", source.Name, "SecretSource.Name returned the wrong value")
	assert.Equal(t, "azure.keyvault", source.PluginSubKey, "SecretSource.PluginSubKey returned the wrong value")
}
