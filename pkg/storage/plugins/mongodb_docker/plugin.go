package mongodb_docker

import (
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"github.com/hashicorp/go-plugin"
	"github.com/mitchellh/mapstructure"
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
func NewPlugin(c *portercontext.Context, rawCfg interface{}) (plugin.Plugin, error) {
	cfg := PluginConfig{
		Port:     "27018",
		Database: "porter",
		Timeout:  10,
	}
	if err := mapstructure.Decode(rawCfg, &cfg); err != nil {
		return nil, fmt.Errorf("error reading plugin configuration: %w", err)
	}

	store := NewStore(c, cfg)
	return pluginstore.NewPlugin(c, store), nil
}
