//go:build integration

package integration

import (
	"bytes"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/pluginstore"
	"github.com/stretchr/testify/require"
)

func TestPlugins_CatchStderr(t *testing.T) {
	c := config.NewTestConfig(t)
	defer c.Close()
	ctx, _, _ := c.SetupIntegrationTest()

	t.Run("plugin throws an error", func(t *testing.T) {
		pluginsPath, _ := c.GetPluginsDir()
		pluginName := "testplugin"

		pluginDir := path.Join(pluginsPath, pluginName)
		err := exec.Command("mkdir", "-p", pluginDir).Run()
		require.NoError(t, err, "could not create plugin dir")

		// testplugin binary will be in bin. refer "test-integration" in Makefile
		binDir := c.TestContext.FindBinDir()
		err = exec.Command("cp", path.Join(binDir, pluginName), pluginDir).Run()
		require.NoError(t, err, "could not copy test binary")

		// Verify that our test plugin will return a predictable error message when run
		var pluginStderr bytes.Buffer
		var pluginStdout bytes.Buffer
		pluginCmd := exec.Command(filepath.Join(pluginDir, pluginName))
		pluginCmd.Stdout = &pluginStdout
		pluginCmd.Stderr = &pluginStderr
		require.NoError(t, pluginCmd.Run(), "error running test plugin standalone")
		require.Equal(t, "i am just test plugin. i don't function\n", pluginStderr.String(), "the test plugin isn't printing the expected message to stderr")
		require.Equal(t, "i am a little teapot\n", pluginStdout.String(), "the test plugin isn't printing the expected message to stdout")

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
		require.EqualError(t, err, `could not connect to the secrets.testplugin.vault plugin: plugin stderr was i am just test plugin. i don't function
: Unrecognized remote plugin message: i am a little teapot

This usually means that the plugin is either invalid or simply
needs to be recompiled to support the latest protocol.`)
	})
}
