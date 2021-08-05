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
	l := NewPluginLoader(c.Config, func(string, interface{}) (plugins.Plugin, error) {
		return nil, nil
	})

	pluginCfg := PluginTypeConfig{
		GetDefaultPluggable: func(c *config.Config) string {
			return c.Data.DefaultStorage
		},
		GetPluggable: func(c *config.Config, name string) (Entry, error) {
			return c.GetStorage(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			return c.Data.DefaultStoragePlugin
		},
	}

	t.Run("internal plugin", func(t *testing.T) {
		c.Data.DefaultStoragePlugin = "mongodb-docker"

		err := l.selectPlugin(pluginCfg)
		require.NoError(t, err, "error selecting plugin")

		assert.Equal(t, &plugins.PluginKey{Binary: "porter", Implementation: "mongodb-docker", IsInternal: true}, l.SelectedPluginKey)
		assert.Nil(t, l.SelectedPluginConfig)
	})

	t.Run("external plugin", func(t *testing.T) {
		c.Data.DefaultStoragePlugin = "azure.blob"

		err := l.selectPlugin(pluginCfg)
		require.NoError(t, err, "error selecting plugin")

		assert.Equal(t, &plugins.PluginKey{Binary: "azure", Implementation: "blob", IsInternal: false}, l.SelectedPluginKey)
		assert.Nil(t, l.SelectedPluginConfig)
	})

	t.Run("configured plugin", func(t *testing.T) {
		c.Data.DefaultStorage = "azure"
		c.Data.CrudStores = []config.CrudStore{
			{
				config.PluginConfig{
					Name:         "azure",
					PluginSubKey: "azure.blob",
					Config: map[string]interface{}{
						"env": "MyAzureConnString",
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
