package pluginstore

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/pkg/errors"
)

var _ secrets.Store = &Store{}

// Store is a plugin-backed source of secrets. It resolves the appropriate
// plugin based on Porter's config and implements the secrets.Store interface
// using the backing plugin.
type Store struct {
	*config.Config
	*secrets.SecretStore
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
		GetDefaultPluggable: func(datastore *config.Data) string {
			return datastore.GetDefaultSecretSource()
		},
		GetPluggable: func(datastore *config.Data, name string) (pluggable.Entry, error) {
			return datastore.GetSecretSource(name)
		},
		GetDefaultPlugin: func(datastore *config.Data) string {
			return datastore.GetSecretsPlugin()
		},
	}
}

func (s *Store) Connect() error {
	if s.SecretStore != nil {
		return nil
	}

	pluginType := NewSecretsPluginConfig()

	l := pluggable.NewPluginLoader(s.Config)
	raw, cleanup, err := l.Load(pluginType)
	if err != nil {
		return err
	}
	s.cleanup = cleanup

	store, ok := raw.(secrets.Store)
	if !ok {
		cleanup()
		return errors.Errorf("the interface exposed by the %s plugin was not secrets.Store", l.SelectedPluginKey)
	}

	s.SecretStore = secrets.NewSecretStore(store)

	return nil
}

func (s *Store) Close() error {
	if s.cleanup != nil {
		s.cleanup()
	}
	s.SecretStore = nil
	return nil
}
