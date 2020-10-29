package parameters

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cnabio/cnab-go/utils/crud"
)

// ItemType is the location in the backing store where parameters are persisted.
const ItemType = "parameters"

// ErrNotFound represents a parameter set not found in storage
var ErrNotFound = errors.New("Parameter set does not exist")

// Store is a persistent store for parameter sets.
type Store struct {
	backingStore crud.ManagedStore
}

// NewParameterStore creates a persistent store for parameter sets using the specified
// backing key-blob store.
func NewParameterStore(store crud.ManagedStore) Store {
	return Store{
		backingStore: store,
	}
}

// GetBackingStore returns the data store behind this credentials store.
func (s Store) GetBackingStore() crud.ManagedStore {
	return s.backingStore
}

// List the names of the stored parameter sets.
func (s Store) List() ([]string, error) {
	return s.backingStore.List(ItemType, "")
}

// Save a parameter set. Any previous version of the parameter set is overwritten.
func (s Store) Save(param ParameterSet) error {
	bytes, err := json.MarshalIndent(param, "", "  ")
	if err != nil {
		return err
	}
	return s.backingStore.Save(ItemType, "", param.Name, bytes)
}

// Read loads the parameter set with the given name from the store.
func (s Store) Read(name string) (ParameterSet, error) {
	bytes, err := s.backingStore.Read(ItemType, name)
	if err != nil {
		if strings.Contains(err.Error(), crud.ErrRecordDoesNotExist.Error()) {
			return ParameterSet{}, ErrNotFound
		}
		return ParameterSet{}, err
	}
	paramset := ParameterSet{}
	err = json.Unmarshal(bytes, &paramset)
	return paramset, err
}

// ReadAll retrieves all the parameter sets.
func (s Store) ReadAll() ([]ParameterSet, error) {
	results, err := s.backingStore.ReadAll(ItemType, "")
	if err != nil {
		return nil, err
	}

	params := make([]ParameterSet, len(results))
	for i, bytes := range results {
		var cs ParameterSet
		err = json.Unmarshal(bytes, &cs)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling parameter set: %v", err)
		}
		params[i] = cs
	}

	return params, nil
}

// Delete deletes a parameter set from the store.
func (s Store) Delete(name string) error {
	return s.backingStore.Delete(ItemType, name)
}
