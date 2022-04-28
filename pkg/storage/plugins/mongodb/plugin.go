package mongodb

import (
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"github.com/hashicorp/go-plugin"
	"github.com/mitchellh/mapstructure"
)

const PluginKey = plugins.PluginInterface + ".porter.mongodb"

var _ plugins.StorageProtocol = Plugin{}

type Plugin struct {
	*Store
}

// PluginConfig are the configuration settings that can be defined for the
// mongodb plugin in porter.yaml
type PluginConfig struct {
	URL     string `mapstructure:"url"`
	Timeout int    `mapstructure:"timeout,omitempty"`
}

func NewPlugin(c *portercontext.Context, rawCfg interface{}) (plugin.Plugin, error) {
	cfg := PluginConfig{
		Timeout: 10,
	}
	if err := mapstructure.Decode(rawCfg, &cfg); err != nil {
		return nil, fmt.Errorf("error reading plugin configuration: %w", err)
	}

	impl := NewStore(c, cfg)
	return pluginstore.NewPlugin(c, impl), nil
}
