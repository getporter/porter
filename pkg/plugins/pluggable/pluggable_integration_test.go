//go:build integration
// +build integration

package pluggable

import (
	"context"
	"os/exec"
	"path"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/require"
)

func TestPlugins_CatchStderr(t *testing.T) {
	c := config.NewTestConfig(t)
	c.SetupIntegrationTest()

	t.Run("plugin throws an error", func(t *testing.T) {
		pluginsPath, _ := c.GetPluginsDir()
		pluginName := "testplugin"

		err := exec.Command("mkdir", "-p", path.Join(pluginsPath, pluginName)).Run()
		require.NoError(t, err, "could not create plugin dir")

		// testplugin binary will be in bin. refer "test-integration" in Makefile
		binDir := c.TestContext.FindBinDir()
		err = exec.Command("cp", path.Join(binDir, pluginName), path.Join(pluginsPath, pluginName)).Run()
		require.NoError(t, err, "could not copy test binary")

		cfg := PluginTypeConfig{
			Interface: secrets.PluginInterface,
			Plugin:    &secrets.Plugin{},
			GetDefaultPluggable: func(c *config.Config) string {
				return c.Data.DefaultSecrets
			},
			GetPluggable: func(c *config.Config, name string) (Entry, error) {
				return c.GetSecretsPlugin(name)
			},
			GetDefaultPlugin: func(c *config.Config) string {
				return c.Data.DefaultSecretsPlugin
			},
		}
		c.Data.DefaultSecrets = "myplugin"
		c.Data.SecretsPlugin = []config.SecretsPlugin{{
			PluginConfig: config.PluginConfig{
				Name:         "myplugin",
				PluginSubKey: "testplugin.vault",
			},
		}}

		ll := NewPluginLoader(c.Config)
		conn, err := ll.Load(context.Background(), cfg)
		conn.Close()
		require.EqualError(t, err, `could not connect to the secrets.testplugin.vault plugin: Unrecognized remote plugin message: 

This usually means that the plugin is either invalid or simply
needs to be recompiled to support the latest protocol.: i am just test plugin. i don't function
`)
	})
}
