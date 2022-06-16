package pluggable

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginLoader_SelectPlugin(t *testing.T) {
	c := config.NewTestConfig(t)
	l := NewPluginLoader(c.Config)

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

		err := l.selectPlugin(context.Background(), pluginCfg)
		require.NoError(t, err, "error selecting plugin")

		assert.Equal(t, &plugins.PluginKey{Binary: "porter", Implementation: "mongodb-docker", IsInternal: true}, l.selectedPluginKey)
		assert.Nil(t, l.selectedPluginConfig)
	})

	t.Run("external plugin", func(t *testing.T) {
		c.Data.DefaultStoragePlugin = "azure.blob"

		err := l.selectPlugin(context.Background(), pluginCfg)
		require.NoError(t, err, "error selecting plugin")

		assert.Equal(t, &plugins.PluginKey{Binary: "azure", Implementation: "blob", IsInternal: false}, l.selectedPluginKey)
		assert.Nil(t, l.selectedPluginConfig)
	})

	t.Run("configured plugin", func(t *testing.T) {
		c.Data.DefaultStorage = "azure"
		c.Data.StoragePlugins = []config.StoragePlugin{
			{
				PluginConfig: config.PluginConfig{
					Name:         "azure",
					PluginSubKey: "azure.blob",
					Config: map[string]interface{}{
						"env": "MyAzureConnString",
					},
				},
			},
		}

		err := l.selectPlugin(context.Background(), pluginCfg)
		require.NoError(t, err, "error selecting plugin")

		assert.Equal(t, &plugins.PluginKey{Binary: "azure", Implementation: "blob", IsInternal: false}, l.selectedPluginKey)
		assert.Equal(t, c.Data.StoragePlugins[0].Config, l.selectedPluginConfig)
	})
}

func TestPluginLoader_IdentifyRecursiveLoad(t *testing.T) {
	// The plugin loader should proactively identify when a plugin is also trying to load another plugin
	// and then return an error to avoid a recursive call that will go up in flames and bork your computer

	ctx := context.Background()
	c := config.NewTestConfig(t)

	// Set Porter's context to indicate we are in an internal plugin right now
	c.IsInternalPlugin = true
	c.InternalPluginKey = "filesystem"

	l := NewPluginLoader(c.Config)

	pluginCfg := PluginTypeConfig{
		GetDefaultPluggable: func(c *config.Config) string {
			return c.Data.DefaultStorage
		},
		GetPluggable: func(c *config.Config, name string) (Entry, error) {
			return c.GetStorage(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			return c.Data.DefaultSecretsPlugin
		},
	}

	conn, err := l.Load(ctx, pluginCfg)
	if conn != nil {
		conn.Close(ctx)
	}
	tests.RequireErrorContains(t, err, "the internal plugin filesystem tried to load the .porter.host plugin")
}
