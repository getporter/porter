// +build integration

package pluggable

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/require"
)

func TestPlugins_CatchStderr(t *testing.T) {
	c := config.NewTestConfig(t)
	c.SetupIntegrationTest()

	t.Run("plugin throws an error", func(t *testing.T) {
		pluginsPath := c.GetPluginsDir()
		pluginName := "testplugin"

		pluginDir := filepath.Join(pluginsPath, pluginName)
		err := c.FileSystem.MkdirAll(pluginDir, 0755)
		require.NoError(t, err, "could not create plugin dir")

		// testplugin binary will be in bin. refer "test-integration" in Makefile
		c.TestContext.AddTestFile(filepath.Join(c.Getenv("PROJECT_ROOT"), "bin", pluginName+pkgmgmt.FileExt),
			filepath.Join(pluginDir, pluginName+pkgmgmt.FileExt))

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
					PluginSubKey: "testplugin.vault",
				},
			}},
		}
		ll := NewPluginLoader(c.Config)
		_, _, err = ll.Load(cfg)
		require.EqualError(t, err, `could not connect to the secrets.testplugin.vault plugin: Unrecognized remote plugin message: 

This usually means that the plugin is either invalid or simply
needs to be recompiled to support the latest protocol.: i am just test plugin. i don't function
`)
	})
}
