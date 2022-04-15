package secrets

import (
	"context"

	"get.porter.sh/porter/pkg/secrets/plugins"
)

var _ Store = PluginAdapter{}

// PluginAdapter converts between the low-level plugins.SecretsProtocol and
// the secrets.Store interface.
type PluginAdapter struct {
	plugin plugins.SecretsPlugin
}

// NewPluginAdapter wraps the specified storage plugin.
func NewPluginAdapter(plugin plugins.SecretsPlugin) PluginAdapter {
	return PluginAdapter{plugin: plugin}
}

func (a PluginAdapter) Connect(ctx context.Context) error {
	return a.plugin.Connect(ctx)
}

func (a PluginAdapter) Close(ctx context.Context) error {
	return a.plugin.Close(ctx)
}

func (a PluginAdapter) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	err := a.Connect(ctx)
	if err != nil {
		return "", err
	}

	return a.plugin.Resolve(keyName, keyValue)
}
