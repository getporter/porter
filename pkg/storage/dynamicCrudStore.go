package storage

import (
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

var _ crud.Store = &DynamicCrudStore{}

type DynamicCrudStoreBuilder func() (crud.Store, func(), error)

// DynamicCrudStore wraps another backing store that is instantiated just in time before each method call.
type DynamicCrudStore struct {
	crudBuilder DynamicCrudStoreBuilder
}

func NewDynamicCrudStore(builder DynamicCrudStoreBuilder) *DynamicCrudStore {
	return &DynamicCrudStore{
		crudBuilder: builder,
	}
}

func (s DynamicCrudStore) init() (crud.Store, func(), error) {
	crud, cleanup, err := s.crudBuilder()
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not dynamically instantiate a backing store")
	}

	if cleanup == nil {
		cleanup = func() {}
	}

	return crud, cleanup, nil
}

func (s DynamicCrudStore) List(itemType string) ([]string, error) {
	store, cleanup, err := s.init()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return store.List(itemType)
}

func (s DynamicCrudStore) Save(itemType string, name string, data []byte) error {
	store, cleanup, err := s.init()
	if err != nil {
		return err
	}
	defer cleanup()

	return store.Save(itemType, name, data)
}

func (s DynamicCrudStore) Read(itemType string, name string) ([]byte, error) {
	store, cleanup, err := s.init()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return store.Read(itemType, name)
}

func (s DynamicCrudStore) Delete(itemType string, name string) error {
	store, cleanup, err := s.init()
	if err != nil {
		return err
	}
	defer cleanup()

	return store.Delete(itemType, name)
}
