package pluginstore

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/storage/crudstore"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

var _ crud.Store = &Store{}

// Store is a plugin backed source of porter home data.
type Store struct {
	*config.Config
	*crud.BackingStore
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
		Interface: crudstore.PluginInterface,
		Plugin:    &crudstore.Plugin{},
		GetDefaultPluggable: func(datastore *config.Data) string {
			return datastore.GetDefaultStorage()
		},
		GetPluggable: func(datastore *config.Data, name string) (pluggable.Entry, error) {
			return datastore.GetStorage(name)
		},
		GetDefaultPlugin: func(datastore *config.Data) string {
			return datastore.GetDefaultStoragePlugin()
		},
	}
}

func (s *Store) Connect() error {
	if s.BackingStore != nil {
		return nil
	}

	pluginType := NewStoragePluginConfig()

	l := pluggable.NewPluginLoader(s.Config)
	raw, cleanup, err := l.Load(pluginType)
	if err != nil {
		return errors.Wrapf(err, "could not load %s plugin", pluginType.Interface)
	}
	s.cleanup = cleanup

	store, ok := raw.(crud.Store)
	if !ok {
		cleanup()
		return errors.Errorf("the interface exposed by the %s plugin was not crud.Store", l.SelectedPluginKey)
	}

	s.BackingStore = crud.NewBackingStore(store)
	s.BackingStore.AutoClose = false

	return nil
}

func (s *Store) Close() error {
	if s.cleanup != nil {
		s.cleanup()
	}
	s.BackingStore = nil
	return nil
}
