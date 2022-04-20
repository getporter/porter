package pluginstore

import (
	"context"

	"get.porter.sh/porter/pkg/config"
	porterplugins "get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	filesystemsecret "get.porter.sh/porter/pkg/secrets/plugins/filesystem"
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

func (s *Store) Create(keyName string, keyValue string, value string) error {
	if err := s.Connect(); err != nil {
		return err
	}

	err := s.plugin.Create(keyName, keyValue, value)
	if errors.Is(err, plugins.ErrNotImplemented) {
		//TODO: add the doc page link once it exists
		return errors.Wrapf(err, `The current secrets plugin does not support persisting secrets. You need to edit your porter configuration file and configure a different secrets plugin.
		
If you are just testing out Porter, and are not working with production secrets, you can edit your config file and set default-storage-plugin to "filesystem" to use the insecure filesystem plugin. Do not use the filesystem plugin for production data.`)
	}

	return err
}

func (s *Store) createInternalPlugin(key string, pluginConfig interface{}) (porterplugins.Plugin, error) {
	if key == host.PluginKey {
		return host.NewPlugin(), nil
	}

	if key == filesystemsecret.PluginKey {
		homeDir, err := s.GetHomeDir()
		if err != nil {
			return nil, err
		}
		cfg := filesystemsecret.NewConfig(s.DebugPlugins, homeDir)

		return filesystemsecret.NewPlugin(cfg)
	}

	return nil, errors.Errorf("unsupported internal secrets plugin specified %s", key)
}

func (s *Store) Connect(ctx context.Context) error {
	if s.plugin != nil {
		return nil
	}

	pluginType := NewSecretsPluginConfig()

	l := pluggable.NewPluginLoader(s.Config, s.createInternalPlugin)
	raw, cleanup, err := l.Load(ctx, pluginType)
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

func (s *Store) Close(ctx context.Context) error {
	if s.cleanup != nil {
		s.cleanup()
	}
	s.plugin = nil
	return nil
}
