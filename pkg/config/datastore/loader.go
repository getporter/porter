package datastore

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// FromFlagsThenEnvVarsThenConfigFile loads data with the following precedence:
// * Flags (highest)
// * Environment variables where --flag is assumed to be PORTER_FLAG
// * Config file (lowest)
func FromFlagsThenEnvVarsThenConfigFile(cmd *cobra.Command) config.DataStoreLoaderFunc {
	return buildDataLoader(func(v *viper.Viper) {
		v.SetEnvPrefix("PORTER")
		v.AutomaticEnv()

		// Apply the configuration file value to the flag when the flag is not set
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			// Environment variables can't have dashes in them, so bind them to their equivalent
			// keys with underscores, e.g. --debug-plugins binds to PORTER_DEBUG_PLUGINS
			if strings.Contains(f.Name, "-") {
				envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
				v.BindEnv(f.Name, fmt.Sprintf("PORTER_%s", envVarSuffix))
			}

			if !f.Changed && v.IsSet(f.Name) {
				val := v.Get(f.Name)
				cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			}
		})
	})
}

// FromConfigFile loads data from the config file only.
func FromConfigFile(cfg *config.Config) error {
	dataloader := buildDataLoader(nil)
	return dataloader(cfg)
}

func buildDataLoader(viperCfg func(v *viper.Viper)) config.DataStoreLoaderFunc {
	return func(cfg *config.Config) error {
		home := cfg.GetHomeDir()

		v := viper.New()
		v.SetFs(cfg.FileSystem)
		v.AddConfigPath(home)
		err := v.ReadInConfig()

		if viperCfg != nil {
			viperCfg(v)
		}

		var data config.Data
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				data = DefaultDataStore()
			} else {
				return errors.Wrapf(err, "error reading config file at %q", v.ConfigFileUsed())
			}
		} else {
			err = v.Unmarshal(&data)
			if err != nil {
				return errors.Wrapf(err, "error unmarshaling config at %q", v.ConfigFileUsed())
			}
		}

		cfg.Data = &data

		return nil
	}
}

// DefaultDataStore used when no config file is found.
func DefaultDataStore() config.Data {
	return config.Data{}
}
