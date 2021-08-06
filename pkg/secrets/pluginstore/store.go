package pluginstore

import (
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
	plugin  plugins.SecretsProtocol
	cleanup func()
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
			return c.GetSecretSource(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			return c.Data.DefaultSecretsPlugin
		},
		ProtocolVersion: 1,
	}
}

func (s *Store) Resolve(keyName string, keyValue string) (string, error) {
	if err := s.Connect(); err != nil {
		return "", err
	}

	return s.plugin.Resolve(keyName, keyValue)
}

func createInternalPlugin(key string, pluginConfig interface{}) (porterplugins.Plugin, error) {
	if key == host.PluginKey {
		return host.NewPlugin(), nil
	}

	return nil, errors.Errorf("unsupported internal secrets plugin specified %s", key)
}

func (s *Store) Connect() error {
	if s.plugin != nil {
		return nil
	}

	pluginType := NewSecretsPluginConfig()

	l := pluggable.NewPluginLoader(s.Config, createInternalPlugin)
	raw, cleanup, err := l.Load(pluginType)
	if err != nil {
		return err
	}
	s.cleanup = cleanup

	store, ok := raw.(plugins.SecretsProtocol)
	if !ok {
		cleanup()
		return errors.Errorf("the interface exposed by the %s plugin was not plugins.SecretsProtocol", l.SelectedPluginKey)
	}

	s.plugin = store
	return nil
}

func (s *Store) Close() error {
	if s.cleanup != nil {
		s.cleanup()
	}
	s.plugin = nil
	return nil
}
