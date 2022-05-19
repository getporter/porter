//go:build integration
// +build integration

package integration

import (
	"context"
	"os/exec"
	"path"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/pluginstore"
	"github.com/stretchr/testify/require"
)

func TestPlugins_CatchStderr(t *testing.T) {
	c := config.NewTestConfig(t)
	c.SetupIntegrationTest()

	t.Run("plugin throws an error", func(t *testing.T) {
		ctx := context.Background()

		pluginsPath, _ := c.GetPluginsDir()
		pluginName := "testplugin"

		err := exec.Command("mkdir", "-p", path.Join(pluginsPath, pluginName)).Run()
		require.NoError(t, err, "could not create plugin dir")

		// testplugin binary will be in bin. refer "test-integration" in Makefile
		binDir := c.TestContext.FindBinDir()
		err = exec.Command("cp", path.Join(binDir, pluginName), path.Join(pluginsPath, pluginName)).Run()
		require.NoError(t, err, "could not copy test binary")

		cfg := pluggable.PluginTypeConfig{
			Interface: plugins.PluginInterface,
			Plugin:    &pluginstore.Plugin{},
			GetDefaultPluggable: func(c *config.Config) string {
				return c.Data.DefaultSecrets
			},
			GetPluggable: func(c *config.Config, name string) (pluggable.Entry, error) {
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

		ll := pluggable.NewPluginLoader(c.Config)
		conn, err := ll.Load(ctx, cfg)
		if conn != nil {
			conn.Close(ctx)
		}
		require.EqualError(t, err, `could not connect to the secrets.testplugin.vault plugin: Unrecognized remote plugin message: 

This usually means that the plugin is either invalid or simply
needs to be recompiled to support the latest protocol.: i am just test plugin. i don't function
`)
	})
}
