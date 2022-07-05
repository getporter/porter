package config

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/tracing"
	"github.com/osteele/liquid"
	"github.com/osteele/liquid/render"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/attribute"
)

var _ DataStoreLoaderFunc = NoopDataLoader

// NoopDataLoader skips loading the datastore.
func NoopDataLoader(_ context.Context, _ *Config, _ map[string]interface{}) error {
	return nil
}

// LoadFromEnvironment loads data with the following precedence:
// * Environment variables where --flag is assumed to be PORTER_FLAG
// * Config file
// * Flag default (lowest)
func LoadFromEnvironment() DataStoreLoaderFunc {
	return LoadFromViper(func(v *viper.Viper) {
		v.SetEnvPrefix("PORTER")
		v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
		v.AutomaticEnv()

		// Bind open telemetry environment variables
		// See https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace
		v.BindEnv("telemetry.endpoint", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
		v.BindEnv("telemetry.protocol", "OTEL_EXPORTER_OTLP_PROTOCOL", "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
		v.BindEnv("telemetry.insecure", "OTEL_EXPORTER_OTLP_INSECURE", "OTEL_EXPORTER_OTLP_TRACES_INSECURE")
		v.BindEnv("telemetry.certificate", "OTEL_EXPORTER_OTLP_CERTIFICATE", "OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE")
		v.BindEnv("telemetry.headers", "OTEL_EXPORTER_OTLP_HEADERS", "OTEL_EXPORTER_OTLP_TRACES_HEADERS")
		v.BindEnv("telemetry.compression", "OTEL_EXPORTER_OTLP_COMPRESSION", "OTEL_EXPORTER_OTLP_TRACES_COMPRESSION")
		v.BindEnv("telemetry.timeout", "OTEL_EXPORTER_OTLP_TIMEOUT", "OTEL_EXPORTER_OTLP_TRACES_TIMEOUT")
	})
}

// LoadFromFilesystem loads data with the following precedence:
// * Config file
// * Flag default (lowest)
// This is used for testing only.
func LoadFromFilesystem() DataStoreLoaderFunc {
	return LoadFromViper(func(v *viper.Viper) {})
}

// LoadFromViper loads data from a configurable viper instance.
func LoadFromViper(viperCfg func(v *viper.Viper)) DataStoreLoaderFunc {
	return func(ctx context.Context, cfg *Config, templateData map[string]interface{}) error {
		home, _ := cfg.GetHomeDir()

		//lint:ignore SA4006 ignore unused context for now
		ctx, log := tracing.StartSpanWithName(ctx, "LoadFromViper", attribute.String("porter.PORTER_HOME", home))
		defer log.EndSpan()

		v := viper.New()
		v.SetFs(cfg.FileSystem)

		// Consider an empty environment variable as "set", so that you can do things like
		// PORTER_DEFAULT_STORAGE="" and have that override what's in the config file.
		v.AllowEmptyEnv(true)

		// Initialize empty config
		err := v.SetDefaultsFrom(cfg.Data)
		if err != nil {
			return log.Error(fmt.Errorf("error initializing configuration data: %w", err))
		}

		if viperCfg != nil {
			viperCfg(v)
		}

		// Find the config file
		v.AddConfigPath(home)

		// Only read the config file if we are running as porter
		// Skip it for internal plugins since we pass the resolved
		// config directly to the plugins
		if !cfg.IsInternalPlugin {
			err = v.ReadInConfig()
			if err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return log.Error(fmt.Errorf("error reading config file: %w", err))
				}
			}
		}

		cfgFile := v.ConfigFileUsed()
		if cfgFile != "" {
			log.SetAttributes(attribute.String("porter.PORTER_CONFIG", cfgFile))

			cfgContents, err := cfg.FileSystem.ReadFile(cfgFile)
			if err != nil {
				return log.Error(fmt.Errorf("error reading config file template: %w", err))
			}

			// Render any template variables used in the config file
			engine := liquid.NewEngine()
			engine.Delims("${", "}", "${%", "%}")
			tmpl, err := engine.ParseTemplate(cfgContents)
			if err != nil {
				return log.Error(fmt.Errorf("error parsing config file as a liquid template:\n%s\n\n: %w", cfgContents, err))
			}

			finalCfg, err := tmpl.Render(templateData)
			if err != nil {
				return log.Error(fmt.Errorf("error rendering config file as a liquid template:\n%s\n\n: %w", cfgContents, err))
			}

			// Remember what variables are used in the template
			// we use this to resolve variables in the second pass over the config file
			if len(cfg.templateVariables) == 0 {
				cfg.templateVariables = listTemplateVariables(tmpl)
			}

			if err := v.ReadConfig(bytes.NewReader(finalCfg)); err != nil {
				return log.Error(fmt.Errorf("error loading configuration file: %w", err))
			}
		}

		if err = v.Unmarshal(&cfg.Data); err != nil {
			log.Error(fmt.Errorf("error unmarshaling viper config as porter config: %w", err))
		}

		cfg.viper = v
		return nil
	}
}

func listTemplateVariables(tmpl *liquid.Template) []string {
	vars := map[string]struct{}{}
	findTemplateVariables(tmpl.GetRoot(), vars)

	results := make([]string, 0, len(vars))
	for v := range vars {
		results = append(results, v)
	}
	sort.Strings(results)

	return results
}

// findTemplateVariables looks at the template's abstract syntax tree (AST)
// and identifies which variables were used
func findTemplateVariables(curNode render.Node, vars map[string]struct{}) {
	switch v := curNode.(type) {
	case *render.SeqNode:
		for _, childNode := range v.Children {
			findTemplateVariables(childNode, vars)
		}
	case *render.ObjectNode:
		vars[v.Args] = struct{}{}
	}
}
