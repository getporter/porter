package notation

import (
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/signing"
	"get.porter.sh/porter/pkg/signing/plugins"
	"get.porter.sh/porter/pkg/signing/pluginstore"
	"github.com/hashicorp/go-plugin"
	"github.com/mitchellh/mapstructure"
)

const PluginKey = plugins.PluginInterface + ".porter.notation"

var _ plugins.SigningProtocol = &Plugin{}

type PluginConfig struct {
	SigningKey string `mapstructure:"key,omitempty"`
}

// Plugin is the plugin wrapper for accessing secrets from a local filesystem.
type Plugin struct {
	signing.Signer
}

func NewPlugin(c *portercontext.Context, rawCfg interface{}) (plugin.Plugin, error) {
	cfg := PluginConfig{}
	if err := mapstructure.Decode(rawCfg, &cfg); err != nil {
		return nil, fmt.Errorf("error reading plugin configuration: %w", err)
	}

	impl := NewSigner(c, cfg)
	return pluginstore.NewPlugin(c, impl), nil
}
