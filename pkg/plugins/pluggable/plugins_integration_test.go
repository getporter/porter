// +build integration

package pluggable

import (
	"os/exec"
	"path"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/require"
)

func TestPlugins_CatchStderr(t *testing.T) {
        home, err := ioutil.TempDir("/tmp", "porter")
	require.NoError(t, err, "could not create porter home directory")

	c := config.NewTestConfig(t)
	c.SetHomeDir(home)
	c.NewCommand = exec.Command

	t.Run("plugin throws an error", func(t *testing.T) {
		pluginsPath, _ := c.GetPluginsDir()
		pluginName := "badplugin"

		err := c.FileSystem.MkdirAll(path.Join(pluginsPath, pluginName), 0777)
		require.NoError(t, err, "could not create plugin dir")
		err = exec.Command("cp", path.Join("testdata", pluginName), path.Join(pluginsPath, pluginName)).Run()
		require.NoError(t, err, "could not copy test binary")

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
		_, _, err = ll.Load(cfg)
		require.EqualError(t, err, `could not connect to the secrets.badplugin.vault plugin: Unrecognized remote plugin message: 

This usually means that the plugin is either invalid or simply
needs to be recompiled to support the latest protocol.: invalid plugin key
`)
	})
}
