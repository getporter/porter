package mock

import (
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/signing"
	"get.porter.sh/porter/pkg/signing/plugins"
	"get.porter.sh/porter/pkg/signing/pluginstore"
	"github.com/hashicorp/go-plugin"
)

const PluginKey = plugins.PluginInterface + ".porter.mock"

var _ plugins.SigningProtocol = &Plugin{}

// Plugin is the plugin wrapper for accessing secrets from a local filesystem.
type Plugin struct {
	signing.Signer
}

func NewPlugin(c *portercontext.Context, rawCfg interface{}) (plugin.Plugin, error) {
	impl := NewSigner()
	return pluginstore.NewPlugin(c, impl), nil
}
