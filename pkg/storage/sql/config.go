package sql

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"get.porter.sh/porter/pkg/config"
)

type PluginConfig struct {
	URL string `mapstructure:"url"`
}

func UnmarshalPluginConfig(rawCfg interface{}) (cfg PluginConfig, err error) {
	err = mapstructure.Decode(rawCfg, &cfg)
	if err != nil {
		err = fmt.Errorf("error reading plugin configuration: %w", err)
	}
	return
}

func IsSQLStore(c *config.Config) (s config.StoragePlugin, ok bool) {
	// TODO add other sql databases
	if c.Data.DefaultStoragePlugin == "postgres" {
		for _, s = range c.Data.StoragePlugins {
			if s.GetPluginSubKey() == "postgres" {
				return s, true
			}
		}
	}
	s, err := c.GetStorage(c.Data.DefaultStorage)
	if err != nil {
		return
	}
	ok = s.GetPluginSubKey() == "postgres"
	return s, ok
}
