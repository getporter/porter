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
		v.AutomaticEnv()
		v.SetEnvPrefix("PORTER")
		v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

		// Apply the configuration file value to the flag when the flag is not set
		flags := cmd.Flags()
		flags.VisitAll(func(f *pflag.Flag) {
			viperKey := f.Name

			// Check if a viper key has been explicitly configured
			if altKey, ok := f.Annotations["viper-key"]; ok {
				if len(altKey) > 0 {
					viperKey = altKey[0]
				}
			}

			if !f.Changed && v.IsSet(viperKey) {
				val := getFlagValue(v, viperKey)
				flags.Set(f.Name, val)
			}
		})
	})
}

func getFlagValue(v *viper.Viper, key string) string {
	val := v.Get(key)

	switch typedValue := val.(type) {
	case []interface{}:
		// slice flags should be set using a,b,c not [a,b,c]
		items := make([]string, len(typedValue))
		for i, item := range typedValue {
			items[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(items, ",")
	default:
		return fmt.Sprintf("%v", val)
	}
}
