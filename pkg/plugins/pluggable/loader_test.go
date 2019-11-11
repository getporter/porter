package pluggable

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginLoader_SelectPlugin(t *testing.T) {
	c := config.NewTestConfig(t)
	l := NewPluginLoader(c.Config)

	pluginCfg := PluginTypeConfig{
		GetDefaultPluggable: func(datastore *config.Data) string {
			return datastore.GetDefaultStorage()
		},
		GetPluggable: func(datastore *config.Data, name string) (Entry, error) {
			return datastore.GetStorage(name)
		},
		GetDefaultPlugin: func(datastore *config.Data) string {
			return datastore.GetStoragePlugin()
		},
	}

	t.Run("internal plugin", func(t *testing.T) {
		c.Data = &config.Data{
			StoragePlugin: "filesystem",
		}

		err := l.selectPlugin(pluginCfg)
		require.NoError(t, err, "error selecting plugin")

		assert.Equal(t, &plugins.PluginKey{Binary: "porter", Implementation: "filesystem", IsInternal: true}, l.SelectedPluginKey)
		assert.Nil(t, l.SelectedPluginConfig)
	})

	t.Run("external plugin", func(t *testing.T) {
		c.Data = &config.Data{
			StoragePlugin: "azure.blob",
		}

		err := l.selectPlugin(pluginCfg)
		require.NoError(t, err, "error selecting plugin")

		assert.Equal(t, &plugins.PluginKey{Binary: "azure", Implementation: "blob", IsInternal: false}, l.SelectedPluginKey)
		assert.Nil(t, l.SelectedPluginConfig)
	})

	t.Run("configured plugin", func(t *testing.T) {
		c.Data = &config.Data{
			DefaultStorage: "azure",
			CrudStores: []config.CrudStore{
				{
					config.PluginConfig{
						Name:         "azure",
						PluginSubKey: "azure.blob",
						Config: map[string]interface{}{
							"env": "MyAzureConnString",
						},
					},
				},
			},
		}

		err := l.selectPlugin(pluginCfg)
		require.NoError(t, err, "error selecting plugin")

		assert.Equal(t, &plugins.PluginKey{Binary: "azure", Implementation: "blob", IsInternal: false}, l.SelectedPluginKey)
		assert.Equal(t, c.Data.CrudStores[0].Config, l.SelectedPluginConfig)
	})
}
