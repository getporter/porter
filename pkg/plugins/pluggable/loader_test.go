package pluggable

import (
	"os/exec"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/secrets"
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
			return datastore.GetDefaultStoragePlugin()
		},
	}

	t.Run("internal plugin", func(t *testing.T) {
		c.Data = &config.Data{
			DefaultStoragePlugin: "filesystem",
		}

		err := l.selectPlugin(pluginCfg)
		require.NoError(t, err, "error selecting plugin")

		assert.Equal(t, &plugins.PluginKey{Binary: "porter", Implementation: "filesystem", IsInternal: true}, l.SelectedPluginKey)
		assert.Nil(t, l.SelectedPluginConfig)
	})

	t.Run("plugin throws an error", func(t *testing.T) {
		porterPath, _ := c.GetHomeDir()
		exec.Command("mkdir", porterPath+"/plugins/badplugin/").Run()
		exec.Command("cp", "testdata/badplugin", porterPath+"/plugins/badplugin/").Run()
		defer exec.Command("rm", "-rf", porterPath+"/plugins/badplugin").Run()

		cfg := PluginTypeConfig{
			Interface: secrets.PluginInterface,
			Plugin:    &secrets.Plugin{},
			GetDefaultPluggable: func(datastore *config.Data) string {
				return datastore.GetDefaultSecretSource()
			},
			GetPluggable: func(datastore *config.Data, name string) (Entry, error) {
				return datastore.GetSecretSource(name)
			},
			GetDefaultPlugin: func(datastore *config.Data) string {
				return datastore.GetDefaultSecretsPlugin()
			},
		}
		c.Data = &config.Data{
			DefaultSecrets: "myplugin",
			SecretSources: []config.SecretSource{{
				PluginConfig: config.PluginConfig{
					Name:         "myplugin",
					PluginSubKey: "badplugin.vault",
				},
			}},
		}
		ll := NewPluginLoader(c.Config)
		_, _, err := ll.Load(cfg)
		require.EqualError(t, err, `could not connect to the secrets.badplugin.vault plugin: Unrecognized remote plugin message: 

This usually means that the plugin is either invalid or simply
needs to be recompiled to support the latest protocol.: invalid plugin key
`)
	})

	t.Run("external plugin", func(t *testing.T) {
		c.Data = &config.Data{
			DefaultStoragePlugin: "azure.blob",
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
