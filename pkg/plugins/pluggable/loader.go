package pluggable

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// PluginStartTimeoutDefault is the default amount of time to wait for a plugin
	// to start. Override with PluginStartTimeoutEnvVar.
	PluginStartTimeoutDefault = 1 * time.Second

	// PluginStopTimeoutDefault is the default amount of time to wait for a plugin
	// to stop (kill). Override with PluginStopTimeoutEnvVar.
	PluginStopTimeoutDefault = 100 * time.Millisecond

	// PluginStartTimeoutEnvVar is the environment variable used to override
	// PluginStartTimeoutDefault.
	PluginStartTimeoutEnvVar = "PORTER_PLUGIN_START_TIMEOUT"

	// PluginStopTimeoutEnvVar is the environment variable used to override
	// PluginStopTimeoutDefault.
	PluginStopTimeoutEnvVar = "PORTER_PLUGIN_STOP_TIMEOUT"
)

// PluginLoader handles finding, configuring and loading porter plugins.
type PluginLoader struct {
	// config is the Porter configuration
	config *config.Config

	// selectedPluginKey is the loaded plugin.
	selectedPluginKey *plugins.PluginKey

	// selectedPluginConfig is the relevant section of the Porter config file containing
	// the plugin's configuration.
	selectedPluginConfig interface{}
}

func NewPluginLoader(c *config.Config) *PluginLoader {
	return &PluginLoader{
		config: c,
	}
}

// Load a plugin, returning the plugin's interface which the caller must then cast to
// the typed interface, a cleanup function to stop the plugin when finished communicating with it,
// and an error if the plugin could not be loaded.
func (l *PluginLoader) Load(ctx context.Context, pluginType PluginTypeConfig) (*PluginConnection, error) {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("plugin-interface", pluginType.Interface),
		attribute.String("requested-protocol-version", fmt.Sprintf("%v", pluginType.ProtocolVersion)))
	defer span.EndSpan()

	err := l.selectPlugin(ctx, pluginType)
	if err != nil {
		return nil, err
	}

	// quick check to detect that we are running as porter, and not a plugin already
	if l.config.IsInternalPlugin {
		err := fmt.Errorf("the internal plugin %s tried to load the %s plugin. Report this error to https://github.com/getporter/porter", l.config.InternalPluginKey, l.selectedPluginKey)
		return nil, span.Error(err)
	}

	span.SetAttributes(attribute.String("plugin-key", l.selectedPluginKey.String()))

	configReader, err := l.readPluginConfig()
	if err != nil {
		return nil, span.Error(err)
	}

	conn := NewPluginConnection(l.config, pluginType, *l.selectedPluginKey)
	if err = conn.Start(ctx, configReader); err != nil {
		return nil, err
	}

	return conn, nil
}

// selectPlugin picks the plugin to use and loads its configuration.
func (l *PluginLoader) selectPlugin(ctx context.Context, cfg PluginTypeConfig) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	l.selectedPluginKey = nil
	l.selectedPluginConfig = nil

	var pluginKey string

	defaultStore := cfg.GetDefaultPluggable(l.config)
	if defaultStore != "" {
		span.SetAttributes(attribute.String("default-plugin", defaultStore))

		is, err := cfg.GetPluggable(l.config, defaultStore)
		if err != nil {
			return span.Error(err)
		}

		pluginKey = is.GetPluginSubKey()
		l.selectedPluginConfig = is.GetConfig()
		if l.selectedPluginConfig == nil {
			span.Debug("No plugin config defined")
		}
	}

	// If there isn't a specific plugin configured for this plugin type, fall back to the default plugin for this type
	if pluginKey == "" {
		pluginKey = cfg.GetDefaultPlugin(l.config)
		span.Debug("Selected default plugin", attribute.String("plugin-key", pluginKey))
	} else {
		span.Debug("Selected configured plugin", attribute.String("plugin-key", pluginKey))
	}

	key, err := plugins.ParsePluginKey(pluginKey)
	if err != nil {
		return span.Error(err)
	}
	l.selectedPluginKey = &key
	l.selectedPluginKey.Interface = cfg.Interface

	return nil
}

func (l *PluginLoader) readPluginConfig() (io.Reader, error) {
	if l.selectedPluginConfig == nil {
		return &bytes.Buffer{}, nil
	}

	b, err := json.Marshal(l.selectedPluginConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal plugin config %#v", l.selectedPluginConfig)
	}

	return bytes.NewBuffer(b), nil
}
