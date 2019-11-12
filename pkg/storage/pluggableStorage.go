package storage

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/claimstore"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

var _ StorageProvider = &PluggableStorage{}

// A sad hack because claim.Store has a method called Store which prevents us from embedding it as a field
type ClaimStore = claim.Store

// PluggableStorage provides access to instance storage (claims) by instantiating plugins that
// implement claim (CRUD) storage.
type PluggableStorage struct {
	*config.Config
	ClaimStore
}

func NewPluggableStorage(c *config.Config) *PluggableStorage {
	l := &PluggableStorage{
		Config: c,
	}

	crud := NewDynamicCrudStore(l.connect)
	// this is silly, we can't embed Store because it has a method named Store...
	l.ClaimStore = claim.NewClaimStore(crud)

	return l
}

// NewPluginTypeConfig for instance storage.
func NewPluginTypeConfig() pluggable.PluginTypeConfig {
	return pluggable.PluginTypeConfig{
		Interface: claimstore.PluginInterface,
		Plugin:    &claimstore.Plugin{},
		GetDefaultPluggable: func(datastore *config.Data) string {
			return datastore.GetDefaultInstanceStore()
		},
		GetPluggable: func(datastore *config.Data, name string) (pluggable.Entry, error) {
			return datastore.GetInstanceStore(name)
		},
		GetDefaultPlugin: func(datastore *config.Data) string {
			return datastore.GetStoragePlugin()
		},
	}
}

func (d *PluggableStorage) connect() (crud.Store, func(), error) {
	pluginType := NewPluginTypeConfig()

	l := pluggable.NewPluginLoader(d.Config)
	raw, cleanup, err := l.Load(pluginType)
	if err != nil {
		return nil, nil, err
	}

	store, ok := raw.(crud.Store)
	if !ok {
		cleanup()
		return nil, nil, errors.Errorf("the interface exposed by the %s plugin was not instance-store.ClaimStore", l.SelectedPluginKey)
	}

	return store, cleanup, nil
}
