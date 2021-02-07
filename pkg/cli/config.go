package cli

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// LoadHierarchicalConfig loads data with the following precedence:
// * User set flag Flags (highest)
// * Environment variables where --flag is assumed to be PORTER_FLAG
// * Config file
// * Flag default (lowest)
func LoadHierarchicalConfig(cmd *cobra.Command) config.DataStoreLoaderFunc {
	return config.LoadFromViper(func(v *viper.Viper) {
		v.SetEnvPrefix("PORTER")
		v.AutomaticEnv()

		// Apply the configuration file value to the flag when the flag is not set
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			viperKey := f.Name

			// Check if a viper key has been explicitly configured
			if altKey, ok := f.Annotations["viper-key"]; ok {
				if len(altKey) > 0 {
					viperKey = altKey[0]
				}
			}

			// Environment variables can't have dashes in them, so bind them to their equivalent
			// keys with underscores, e.g. --debug-plugins binds to PORTER_DEBUG_PLUGINS
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(viperKey, "-", "_"))
			v.BindEnv(viperKey, fmt.Sprintf("PORTER_%s", envVarSuffix))

			if !f.Changed && v.IsSet(viperKey) {
				val := v.Get(viperKey)
				cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			}
		})
	})
}
