package pluginstore

import (
	"get.porter.sh/porter/pkg/config"
	porterplugins "get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

var _ plugins.StoragePlugin = &Store{}

// Store is a plugin-backed source of storage. It resolves the appropriate
// plugin based on Porter's config and implements the plugins.StorageProtocol interface
// using the backing plugin.
//
// Connects just-in-time, but you must call Close to release resources.
type Store struct {
	*config.Config
	plugin  plugins.StorageProtocol
	cleanup func()
}

func NewStore(c *config.Config) *Store {
	return &Store{
		Config: c,
	}
}

// NewStoragePluginConfig for porter home storage.
func NewStoragePluginConfig() pluggable.PluginTypeConfig {
	return pluggable.PluginTypeConfig{
		Interface: plugins.PluginInterface,
		Plugin:    &Plugin{},
		GetDefaultPluggable: func(c *config.Config) string {
			return c.Data.DefaultStorage
		},
		GetPluggable: func(c *config.Config, name string) (pluggable.Entry, error) {
			return c.GetStorage(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			return c.Data.DefaultStoragePlugin
		},
		ProtocolVersion: 2,
	}
}

func (s *Store) createInternalPlugin(key string, pluginConfig interface{}) (porterplugins.Plugin, error) {
	switch key {
	case mongodb_docker.PluginKey:
		return mongodb_docker.NewPlugin(s.Context, pluginConfig)
	case mongodb.PluginKey:
		return mongodb.NewPlugin(s.Context, pluginConfig)
	default:
		return nil, errors.Errorf("unsupported internal storage plugin specified %s", key)
	}
}

func (s *Store) Connect() error {
	if s.plugin != nil {
		return nil
	}

	pluginType := NewStoragePluginConfig()

	l := pluggable.NewPluginLoader(s.Config, s.createInternalPlugin)
	raw, cleanup, err := l.Load(pluginType)
	if err != nil {
		return errors.Wrapf(err, "could not load %s plugin", pluginType.Interface)
	}
	s.cleanup = cleanup

	store, ok := raw.(plugins.StorageProtocol)
	if !ok {
		cleanup()
		return errors.Errorf("the interface exposed by the %s plugin was not plugins.StorageProtocol", l.SelectedPluginKey)
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

func (s *Store) EnsureIndex(opts plugins.EnsureIndexOptions) error {
	if err := s.Connect(); err != nil {
		return err
	}

	return s.plugin.EnsureIndex(opts)
}

func (s *Store) Aggregate(opts plugins.AggregateOptions) ([]bson.Raw, error) {
	if err := s.Connect(); err != nil {
		return nil, err
	}

	return s.plugin.Aggregate(opts)
}

func (s *Store) Count(opts plugins.CountOptions) (int64, error) {
	if err := s.Connect(); err != nil {
		return 0, err
	}

	return s.plugin.Count(opts)
}

func (s *Store) Find(opts plugins.FindOptions) ([]bson.Raw, error) {
	if err := s.Connect(); err != nil {
		return nil, err
	}

	return s.plugin.Find(opts)
}

func (s *Store) Insert(opts plugins.InsertOptions) error {
	if err := s.Connect(); err != nil {
		return err
	}

	return s.plugin.Insert(opts)
}

func (s *Store) Patch(opts plugins.PatchOptions) error {
	if err := s.Connect(); err != nil {
		return err
	}

	return s.plugin.Patch(opts)
}

func (s *Store) Remove(opts plugins.RemoveOptions) error {
	if err := s.Connect(); err != nil {
		return err
	}

	return s.plugin.Remove(opts)
}

func (s *Store) Update(opts plugins.UpdateOptions) error {
	if err := s.Connect(); err != nil {
		return err
	}

	return s.plugin.Update(opts)
}
