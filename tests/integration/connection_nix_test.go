//go:build integration && !windows

package integration

import (
	"context"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/portercontext"
	secretsplugins "get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/pluginstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestPluginConnectionForLeaks(t *testing.T) {
	// We are only running this test on linux/darwin because windows doesn't have SIGCONT
	ctx := context.Background()

	c := config.NewTestConfig(t)
	defer c.Close()
	c.SetupIntegrationTest()
	c.ConfigureLogging(ctx, portercontext.LogConfiguration{
		LogLevel:       zapcore.DebugLevel,
		StructuredLogs: true,
	})
	ctx, log := c.StartRootSpan(ctx, t.Name())
	defer log.Close()

	pluginKey := plugins.PluginKey{
		Interface:      secretsplugins.PluginInterface,
		Binary:         "porter",
		Implementation: "host",
		IsInternal:     true,
	}
	pluginType := pluginstore.NewSecretsPluginConfig()
	conn := pluggable.NewPluginConnection(c.Config, pluginType, pluginKey)
	defer conn.Close(ctx)

	err := conn.Start(ctx, strings.NewReader(""))
	require.NoError(t, err, "failed to start the plugin")
	assert.True(t, conn.IsPluginRunning(), "the plugin did not start")

	err = conn.Close(ctx)
	require.NoError(t, err, "failed to close the plugin")
	assert.False(t, conn.IsPluginRunning(), "the plugin connection was leaked")
}
