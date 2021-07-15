package filesystem

import (
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/storage/plugins"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// PluginKey is the identifier of the internal filesystem plugin.
const PluginKey = plugins.PluginInterface + ".porter.filesystem"

// PluginConfig supported by the filesystem plugin as defined in porter.yaml
type PluginConfig struct {
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
}

// NewPlugin creates an instance of the storage.porter.filesystem plugin
func NewPlugin(cxt *context.Context, pluginConfig interface{}) (plugins.StoragePlugin, error) {
	cfg := PluginConfig{
		Port:     "27018",
		Database: "porter",
	}
	if err := mapstructure.Decode(pluginConfig, &cfg); err != nil {
		return nil, errors.Wrapf(err, "error decoding %s plugin config from %#v", PluginKey, pluginConfig)
	}

	return NewStore(cxt, cfg), nil
}
