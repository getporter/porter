package pluginstore

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/pkg/errors"
)

var _ plugins.SecretsProtocol = &Store{}

// Store is a plugin-backed source of secrets. It resolves the appropriate
// plugin based on Porter's config and implements the plugins.SecretsProtocol interface
// using the backing plugin.
//
// Connects just-in-time, but you must call Close to release resources.
type Store struct {
	*config.Config
	plugin plugins.SecretsProtocol
	conn   *pluggable.PluginConnection
}

func NewStore(c *config.Config) *Store {
	return &Store{
		Config: c,
	}
}

// NewSecretsPluginConfig for secret sources.
func NewSecretsPluginConfig() pluggable.PluginTypeConfig {
	return pluggable.PluginTypeConfig{
		Interface: plugins.PluginInterface,
		Plugin:    &Plugin{},
		GetDefaultPluggable: func(c *config.Config) string {
			return c.Data.DefaultSecrets
		},
		GetPluggable: func(c *config.Config, name string) (pluggable.Entry, error) {
			return c.GetSecretsPlugin(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			return c.Data.DefaultSecretsPlugin
		},
		ProtocolVersion: plugins.PluginProtocolVersion,
	}
}

func (s *Store) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return "", err
	}

	value, err := s.plugin.Resolve(ctx, keyName, keyValue)
	if err != nil {
		return "", span.Error(err)
	}
	return value, nil
}

func (s *Store) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	err := s.plugin.Create(ctx, keyName, keyValue, value)
	if errors.Is(err, plugins.ErrNotImplemented) {
		//TODO: add the doc page link once it exists
		return span.Error(fmt.Errorf(`the current secrets plugin does not support persisting secrets. You need to edit your porter configuration file and configure a different secrets plugin.
		
If you are just testing out Porter, and are not working with production secrets, you can edit your config file and set default-storage-plugin to "filesystem" to use the insecure filesystem plugin. Do not use the filesystem plugin for production data.: %w`, err))
	}

	return span.Error(err)
}

// Connect initializes the plugin for use.
// The plugin itself is responsible for ensuring it was called.
// Close is called automatically when the plugin is used by Porter.
func (s *Store) Connect(ctx context.Context) error {
	if s.plugin != nil {
		return nil
	}

	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	pluginType := NewSecretsPluginConfig()

	l := pluggable.NewPluginLoader(s.Config)
	conn, err := l.Load(ctx, pluginType)
	if err != nil {
		return span.Error(err)
	}
	s.conn = conn

	store, ok := conn.GetClient().(plugins.SecretsProtocol)
	if !ok {
		conn.Close(ctx)
		return span.Error(fmt.Errorf("the interface (%T) exposed by the %s plugin was not plugins.SecretsProtocol", conn.GetClient(), conn))
	}
	s.plugin = store

	return nil
}

func (s *Store) Close() error {
	if s.conn != nil {
		s.conn.Close(context.Background())
		s.conn = nil
	}
	return nil
}
