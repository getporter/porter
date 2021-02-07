package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// LoadFromConfigFile loads data from the config file only.
func LoadFromConfigFile(cfg *Config) error {
	dataloader := LoadFromViper(nil)
	return dataloader(cfg)
}

// LoadFromViper loads data from a configurable viper instance.
func LoadFromViper(viperCfg func(v *viper.Viper)) DataStoreLoaderFunc {
	return func(cfg *Config) error {
		home, _ := cfg.GetHomeDir()

		v := viper.New()
		v.SetFs(cfg.FileSystem)
		v.AddConfigPath(home)
		err := v.ReadInConfig()

		if viperCfg != nil {
			viperCfg(v)
		}

		var data Data
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				data = DefaultDataStore()
			} else {
				return errors.Wrapf(err, "error reading config file at %q", v.ConfigFileUsed())
			}
		}

		err = v.Unmarshal(&data)
		if err != nil {
			return errors.Wrapf(err, "error unmarshaling config at %q", v.ConfigFileUsed())
		}

		cfg.Data = &data

		return nil
	}
}

// DefaultDataStore used when no config file is found.
func DefaultDataStore() Data {
	return Data{}
}
