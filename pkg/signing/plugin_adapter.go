package signing

import (
	"context"
	"io"

	"get.porter.sh/porter/pkg/signing/plugins"
)

var _ Signer = PluginAdapter{}

// PluginAdapter converts between the low-level plugins.SigningProtocol and
// the signing.Signer interface.
type PluginAdapter struct {
	plugin plugins.SigningProtocol
}

// NewPluginAdapter wraps the specified storage plugin.
func NewPluginAdapter(plugin plugins.SigningProtocol) PluginAdapter {
	return PluginAdapter{plugin: plugin}
}

func (a PluginAdapter) Close() error {
	if closer, ok := a.plugin.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (a PluginAdapter) Sign(ctx context.Context, ref string) error {
	return a.plugin.Sign(ctx, ref)
}

func (a PluginAdapter) Verify(ctx context.Context, ref string) error {
	return a.plugin.Verify(ctx, ref)
}

func (a PluginAdapter) Connect(ctx context.Context) error {
	return a.plugin.Connect(ctx)
}
