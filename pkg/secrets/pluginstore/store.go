package pluginstore

import (
	"context"

	"get.porter.sh/porter/pkg/config"
	porterplugins "get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/plugins/host"
	"github.com/pkg/errors"
)

var _ plugins.SecretsPlugin = &Store{}

// Store is a plugin-backed source of secrets. It resolves the appropriate
// plugin based on Porter's config and implements the plugins.SecretsProtocol interface
// using the backing plugin.
//
// Connects just-in-time, but you must call Close to release resources.
type Store struct {
	*config.Config
	plugin plugins.SecretsProtocol
	conn   pluggable.PluginConnection
}

func NewStore(c *config.Config) *Store {
	return &Store{
		Config: c,
	}
}

// NewSecretsPluginConfig for secret sources.
func NewSecretsPluginConfig() pluggable.PluginTypeConfig {
	return pluggable.PluginTypeConfig{
		Interface: secrets.PluginInterface,
		Plugin:    &secrets.Plugin{},
		GetDefaultPluggable: func(c *config.Config) string {
			return c.Data.DefaultSecrets
		},
		GetPluggable: func(c *config.Config, name string) (pluggable.Entry, error) {
			return c.GetSecretsPlugin(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			return c.Data.DefaultSecretsPlugin
		},
		ProtocolVersion: 1,
	}
}

func (s *Store) Resolve(keyName string, keyValue string) (string, error) {
	if err := s.Connect(context.Background()); err != nil {
		return "", err
	}

	return s.plugin.Resolve(keyName, keyValue)
}

func createInternalPlugin(ctx context.Context, key string, pluginConfig interface{}) (porterplugins.Plugin, error) {
	if key == host.PluginKey {
		return host.NewPlugin(ctx), nil
	}

	return nil, errors.Errorf("unsupported internal secrets plugin specified %s", key)
}

func (s *Store) Connect(ctx context.Context) error {
	if s.plugin != nil {
		return nil
	}

	pluginType := NewSecretsPluginConfig()

	l := pluggable.NewPluginLoader(s.Config, createInternalPlugin)
	conn, err := l.Load(ctx, pluginType)
	if err != nil {
		return err
	}
	s.conn = conn

	store, ok := conn.Client.(plugins.SecretsProtocol)
	if !ok {
		conn.Close()
		return errors.Errorf("the interface exposed by the %s plugin was not plugins.SecretsProtocol", l.SelectedPluginKey)
	}
	s.plugin = store

	return nil
}

func (s *Store) Close(ctx context.Context) error {
	s.conn.Close()
	s.plugin = nil
	return nil
}
