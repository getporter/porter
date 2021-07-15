package mongodb

import (
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/storage/plugins"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const PluginKey = plugins.PluginInterface + ".porter.mongodb"

// PluginConfig are the configuration settings that can be defined for the
// mongodb plugin in porter.yaml
type PluginConfig struct {
	URL string `mapstructure:"url"`
}

func NewPlugin(cxt *context.Context, pluginConfig interface{}) (plugins.StoragePlugin, error) {
	cfg := PluginConfig{}
	if err := mapstructure.Decode(pluginConfig, &cfg); err != nil {
		return nil, errors.Wrapf(err, "error decoding %s plugin config from %#v", PluginKey, pluginConfig)
	}

	return NewStore(cxt, cfg), nil
}
