package sbom

import (
	"context"
	"io"

	"get.porter.sh/porter/pkg/sbom/plugins"
)

var _ SBOMGenerator = PluginAdapter{}

// PluginAdapter converts between the low-level plugins.SigningProtocol and
// the signing.SBOMGenerator interface.
type PluginAdapter struct {
	plugin plugins.SBOMGeneratorProtocol
}

// NewPluginAdapter wraps the specified storage plugin.
func NewPluginAdapter(plugin plugins.SBOMGeneratorProtocol) PluginAdapter {
	return PluginAdapter{plugin: plugin}
}

func (a PluginAdapter) Generate(ctx context.Context, bundleRef string, sbomPath string, insecureRegistry bool) error {
	return a.plugin.Generate(ctx, bundleRef, sbomPath, insecureRegistry)
}

func (a PluginAdapter) Close() error {
	if closer, ok := a.plugin.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (a PluginAdapter) Connect(ctx context.Context) error {
	return a.plugin.Connect(ctx)
}
