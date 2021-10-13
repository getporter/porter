package config

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var _ DataStoreLoaderFunc = NoopDataLoader

// NoopDataLoader skips loading the datastore.
func NoopDataLoader(_ *Config) error {
	return nil
}

// LoadHierarchicalConfig loads data with the following precedence:
// * User set flag Flags (highest)
// * Environment variables where --flag is assumed to be PORTER_FLAG
// * Config file
// * Flag default (lowest)
func LoadFromEnvironment() DataStoreLoaderFunc {
	return LoadFromViper(func(v *viper.Viper) {
		v.SetEnvPrefix("PORTER")
		v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		v.AutomaticEnv()
	})
}

// LoadFromViper loads data from a configurable viper instance.
func LoadFromViper(viperCfg func(v *viper.Viper)) DataStoreLoaderFunc {
	return func(cfg *Config) error {
		home, _ := cfg.GetHomeDir()

		v := viper.New()
		v.SetFs(cfg.FileSystem)

		// Consider an empty environment variable as "set", so that you can do things like
		// PORTER_DEFAULT_STORAGE="" and have that override what's in the config file.
		v.AllowEmptyEnv(true)

		// Initialize empty config
		err := v.SetDefaultsFrom(cfg.Data)
		if err != nil {
			return err
		}

		if viperCfg != nil {
			viperCfg(v)
		}

		// Try to read config
		v.AddConfigPath(home)
		if cfg.Debug {
			fmt.Fprintln(cfg.Err, "detecting Porter config in", home)
		}
		err = v.ReadInConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return errors.Wrapf(err, "error reading config file at %q", v.ConfigFileUsed())
			}
		}

		if cfg.Debug {
			fmt.Fprintln(cfg.Err, "loaded Porter config from", v.ConfigFileUsed())
		}
		err = v.Unmarshal(&cfg.Data)
		return errors.Wrapf(err, "error unmarshaling config at %q", v.ConfigFileUsed())
	}
}
