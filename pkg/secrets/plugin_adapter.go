package secrets

import (
	"context"
	"io"

	"get.porter.sh/porter/pkg/secrets/plugins"
)

var _ Store = PluginAdapter{}

// PluginAdapter converts between the low-level plugins.SecretsProtocol and
// the secrets.Store interface.
type PluginAdapter struct {
	plugin plugins.SecretsProtocol
}

// NewPluginAdapter wraps the specified storage plugin.
func NewPluginAdapter(plugin plugins.SecretsProtocol) PluginAdapter {
	return PluginAdapter{plugin: plugin}
}

func (a PluginAdapter) Close() error {
	if closer, ok := a.plugin.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (a PluginAdapter) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	// Instead of calling out to a plugin, resolve the value from Porter's database
	// This supports bundle workflows where we are sourcing data from other runs, e.g. passing a connection string from a dependency to another bundle
	if keyName == "porter" {

	}

	return a.plugin.Resolve(ctx, keyName, keyValue)
}

func (a PluginAdapter) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	return a.plugin.Create(ctx, keyName, keyValue, value)
}
