package mongodb_docker

import (
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// PluginKey is the identifier of the internal mongodb run in docker plugin.
const PluginKey = plugins.PluginInterface + ".porter.mongodb-docker"

// PluginConfig supported by the mongodb-docker plugin as defined in porter.yaml
type PluginConfig struct {
	Port     string `mapstructure:"port,omitempty"`
	Database string `mapstructure:"database,omitempty"`

	// Timeout in seconds
	Timeout int `mapstructure:"timeout,omitempty"`
}

// NewPlugin creates an instance of the storage.porter.mongodb-docker plugin
func NewPlugin(cxt *portercontext.Context, pluginConfig interface{}) (plugins.StoragePlugin, error) {
	cfg := PluginConfig{
		Port:     "27018",
		Database: "porter",
		Timeout:  10,
	}
	if err := mapstructure.Decode(pluginConfig, &cfg); err != nil {
		return nil, errors.Wrapf(err, "error decoding %s plugin config from %#v", PluginKey, pluginConfig)
	}

	return NewStore(cxt, cfg), nil
}
