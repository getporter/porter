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

			if f.Changed {
				// Apply the flag to viper
				v.Set(viperKey, getViperValue(flags, f))
			} else if v.IsSet(viperKey) {
				// Apply viper to the flag
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

func getViperValue(flags *pflag.FlagSet, f *pflag.Flag) interface{} {
	var out interface{}
	var err error

	// This is not an exhaustive list, if we need more types supported, it'll panic, and then we can add it.
	flagType := f.Value.Type()
	switch flagType {
	case "int":
		out, err = flags.GetInt(f.Name)
	case "int64":
		out, err = flags.GetInt64(f.Name)
	case "string":
		out, err = flags.GetString(f.Name)
	case "bool":
		out, err = flags.GetBool(f.Name)
	case "stringSlice":
		out, err = flags.GetStringSlice(f.Name)
	case "stringArray":
		out, err = flags.GetStringArray(f.Name)
	default:
		panic(fmt.Errorf("unsupported type for conversion between flag %s and viper configuration: %T", f.Name, flagType))
	}

	if err != nil {
		panic(fmt.Errorf("error parsing config key %s as %T: %w", f.Name, flagType, err))
	}

	return out
}
