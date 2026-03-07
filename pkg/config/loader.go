package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/tracing"
	"github.com/jeremywohl/flatten"
	"github.com/mitchellh/mapstructure"
	"github.com/osteele/liquid"
	"github.com/osteele/liquid/render"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/attribute"
	"gopkg.in/yaml.v3"
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
	return LoadFromViper(BindViperToEnvironmentVariables, nil)
}

func BindViperToEnvironmentVariables(v *viper.Viper) {
	v.SetEnvPrefix("PORTER")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	// Bind open telemetry environment variables
	// See https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace
	var err error
	if err = v.BindEnv("telemetry.endpoint", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"); err != nil {
		_ = errors.Unwrap(err)
	}
	if err = v.BindEnv("telemetry.protocol", "OTEL_EXPORTER_OTLP_PROTOCOL", "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"); err != nil {
		_ = errors.Unwrap(err)
	}
	if err = v.BindEnv("telemetry.insecure", "OTEL_EXPORTER_OTLP_INSECURE", "OTEL_EXPORTER_OTLP_TRACES_INSECURE"); err != nil {
		_ = errors.Unwrap(err)
	}
	if err = v.BindEnv("telemetry.certificate", "OTEL_EXPORTER_OTLP_CERTIFICATE", "OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE"); err != nil {
		_ = errors.Unwrap(err)
	}
	if err = v.BindEnv("telemetry.headers", "OTEL_EXPORTER_OTLP_HEADERS", "OTEL_EXPORTER_OTLP_TRACES_HEADERS"); err != nil {
		_ = errors.Unwrap(err)
	}
	if err = v.BindEnv("telemetry.compression", "OTEL_EXPORTER_OTLP_COMPRESSION", "OTEL_EXPORTER_OTLP_TRACES_COMPRESSION"); err != nil {
		_ = errors.Unwrap(err)
	}
	if err = v.BindEnv("telemetry.timeout", "OTEL_EXPORTER_OTLP_TIMEOUT", "OTEL_EXPORTER_OTLP_TRACES_TIMEOUT"); err != nil {
		_ = errors.Unwrap(err)
	}
}

// LoadFromFilesystem loads data with the following precedence:
// * Config file
// * Flag default (lowest)
// This is used for testing only.
func LoadFromFilesystem() DataStoreLoaderFunc {
	return LoadFromViper(nil, nil)
}

// LoadFromViper loads data from a configurable viper instance.
func LoadFromViper(viperCfg func(v *viper.Viper), cobraCfg func(v *viper.Viper)) DataStoreLoaderFunc {
	return func(ctx context.Context, cfg *Config, templateData map[string]interface{}) error {
		home, _ := cfg.GetHomeDir()

		_, log := tracing.StartSpanWithName(ctx, "LoadFromViper", attribute.String("porter.PORTER_HOME", home))
		defer log.EndSpan()

		v := viper.New()
		v.SetFs(cfg.FileSystem)

		// Consider an empty environment variable as "set", so that you can do things like
		// PORTER_DEFAULT_STORAGE="" and have that override what's in the config file.
		v.AllowEmptyEnv(true)

		// Initialize empty config
		// 2024-12-23: This is still needed, otherwise TestLegacyPluginAdapter fails.
		err := setDefaultsFrom(v, cfg.Data)
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

		var (
			cfgContents  []byte
			cfgEngine    *liquid.Engine
			cfgTmpl      *liquid.Template
			preRenderMap map[string]interface{}
		)

		cfgFile := v.ConfigFileUsed()
		if cfgFile != "" {
			log.SetAttributes(attribute.String("porter.PORTER_CONFIG", cfgFile))

			cfgContents, err = cfg.FileSystem.ReadFile(cfgFile)
			if err != nil {
				return log.Error(fmt.Errorf("error reading config file template: %w", err))
			}

			// Render any template variables used in the config file
			cfgEngine = liquid.NewEngine()
			cfgEngine.Delims("${", "}", "${%", "%}")
			cfgTmpl, err = cfgEngine.ParseTemplate(cfgContents)
			if err != nil {
				return log.Error(fmt.Errorf("error parsing config file as a liquid template:\n%s\n\n: %w", cfgContents, err))
			}

			// Snapshot viper settings before rendering so that template
			// syntax (e.g. ${secret.X}) is preserved as plain strings.
			// This snapshot is used later to detect which secrets belong
			// to the selected context without re-parsing the raw bytes.
			preRenderMap = v.AllSettings()

			finalCfg, err := cfgTmpl.Render(templateData)
			if err != nil {
				return log.Error(fmt.Errorf("error rendering config file as a liquid template:\n%s\n\n: %w", cfgContents, err))
			}

			if err := v.ReadConfig(bytes.NewReader(finalCfg)); err != nil {
				return log.Error(fmt.Errorf("error loading configuration file: %w", err))
			}
		}

		// Porter can be used through the CLI, in which case give it a chance to hook up cobra command flags to viper
		if cobraCfg != nil {
			cobraCfg(v)
		}

		// Read PORTER_CONTEXT env var as fallback when --context flag wasn't set
		if cfg.ContextName == "" {
			cfg.ContextName = v.GetString("context")
		}

		rawMap := v.AllSettings()
		if _, isMultiContext := rawMap["schemaversion"]; isMultiContext {
			// New multi-context format: extract the selected context's config
			// sub-map and load it into a fresh viper so that env vars and
			// cobra flags can still override individual values.
			selected := cfg.ContextName
			if selected == "" {
				// Fall back to current-context from the file, then "default"
				if cc, _ := rawMap["current-context"].(string); cc != "" {
					selected = cc
				} else {
					selected = "default"
				}
			}

			// Collect template variables only from the selected context so
			// secrets in other contexts are never resolved unnecessarily.
			// preRenderMap holds viper settings captured before liquid
			// rendering, so ${secret.X} values are still plain strings.
			if len(cfg.templateVariables) == 0 && cfgEngine != nil && preRenderMap != nil {
				vars, err := listContextTemplateVariables(cfgEngine, preRenderMap, selected)
				if err != nil {
					return log.Error(fmt.Errorf("error scanning config template variables: %w", err))
				}
				cfg.templateVariables = vars
			}

			contextConfigMap, err := extractContextConfig(rawMap, selected)
			if err != nil {
				return log.Error(err)
			}

			ctxViper := viper.New()
			ctxViper.SetFs(cfg.FileSystem)
			if err := setDefaultsFrom(ctxViper, cfg.Data); err != nil {
				return log.Error(err)
			}
			if err := ctxViper.MergeConfigMap(contextConfigMap); err != nil {
				return log.Error(fmt.Errorf("error merging context config: %w", err))
			}
			if viperCfg != nil {
				viperCfg(ctxViper)
			}
			if cobraCfg != nil {
				cobraCfg(ctxViper)
			}
			if err := ctxViper.Unmarshal(&cfg.Data); err != nil {
				return log.Error(fmt.Errorf("error loading context config: %w", err))
			}
			cfg.viper = ctxViper
		} else {
			// Legacy flat format — existing path unchanged.
			if cfg.ContextName != "" {
				return log.Error(fmt.Errorf("--context/PORTER_CONTEXT requires a versioned config file; add schemaVersion: %q and wrap settings under contexts", ConfigSchemaVersion))
			}
			// Collect template variables from the full config file.
			if len(cfg.templateVariables) == 0 && cfgTmpl != nil {
				cfg.templateVariables = listTemplateVariables(cfgTmpl)
			}
			if err := v.Unmarshal(&cfg.Data); err != nil {
				return fmt.Errorf("error unmarshaling viper config as porter config: %w", err)
			}
			cfg.viper = v
		}

		return nil
	}
}

// extractContextConfig finds the named context in the raw viper settings map
// and returns its "config" sub-map. Returns an empty map (not an error) when
// the context exists but has no config block. Returns an error when the
// contexts key is missing entirely or the named context is not found.
func extractContextConfig(rawMap map[string]interface{}, name string) (map[string]interface{}, error) {
	contextsRaw, ok := rawMap["contexts"]
	if !ok {
		return nil, fmt.Errorf("versioned config file is missing required 'contexts' key")
	}
	contexts, ok := contextsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid 'contexts' field in config file")
	}

	var availableNames []string
	for _, c := range contexts {
		ctxMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		ctxName, _ := ctxMap["name"].(string)
		if ctxName == "" {
			continue
		}
		availableNames = append(availableNames, ctxName)
		if ctxName == name {
			if configMap, ok := ctxMap["config"].(map[string]interface{}); ok {
				return configMap, nil
			}
			return map[string]interface{}{}, nil
		}
	}

	return nil, fmt.Errorf("context %q not found in config file; available: %s",
		name, strings.Join(availableNames, ", "))
}

func setDefaultsFrom(v *viper.Viper, val interface{}) error {
	var tmp map[string]interface{}
	err := mapstructure.Decode(val, &tmp)
	if err != nil {
		return fmt.Errorf("error decoding configuration from struct: %v", err)
	}

	defaults, err := flatten.Flatten(tmp, "", flatten.DotStyle)
	if err != nil {
		return fmt.Errorf("error flattening default configuration from struct: %v", err)
	}
	for defaultKey, defaultValue := range defaults {
		v.SetDefault(defaultKey, defaultValue)
	}
	return nil
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

// listContextTemplateVariables returns template variables referenced only
// within the named context's config block. rawMap must be a pre-render
// viper snapshot (captured before liquid rendering) so that template
// syntax such as ${secret.X} is still present as plain string values.
// Using the already-parsed viper map makes this format-agnostic: TOML,
// JSON, HCL and YAML configs are all handled correctly by viper before
// this function is called.
func listContextTemplateVariables(engine *liquid.Engine, rawMap map[string]interface{}, contextName string) ([]string, error) {
	configBlock, err := extractContextConfig(rawMap, contextName)
	if err != nil {
		// Context not found in pre-render map; the error will be surfaced
		// again by the extractContextConfig call in the main load path.
		return nil, nil
	}
	configYAML, err := yaml.Marshal(configBlock)
	if err != nil {
		return nil, err
	}
	tmpl, err := engine.ParseTemplate(configYAML)
	if err != nil {
		return nil, err
	}
	return listTemplateVariables(tmpl), nil
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
