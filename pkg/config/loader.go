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

		// Bind open telemetry environment variables
		// See https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace
		v.BindEnv("trace.endpoint", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
		v.BindEnv("trace.protocol", "OTEL_EXPORTER_OTLP_PROTOCOL", "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
		v.BindEnv("trace.insecure", "OTEL_EXPORTER_OTLP_INSECURE", "OTEL_EXPORTER_OTLP_TRACES_INSECURE")
		v.BindEnv("trace.certificate", "OTEL_EXPORTER_OTLP_CERTIFICATE", "OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE")
		v.BindEnv("trace.headers", "OTEL_EXPORTER_OTLP_HEADERS", "OTEL_EXPORTER_OTLP_TRACES_HEADERS")
		v.BindEnv("trace.compression", "OTEL_EXPORTER_OTLP_COMPRESSION", "OTEL_EXPORTER_OTLP_TRACES_COMPRESSION")
		v.BindEnv("trace.timeout", "OTEL_EXPORTER_OTLP_TIMEOUT", "OTEL_EXPORTER_OTLP_TRACES_TIMEOUT")
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
