package credentials

import (
	"encoding/json"
	"errors"

	"github.com/deislabs/cnab-go/utils/crud"
)

const ItemType = "credentials"

// ErrNotFound represents a credential set not found in storage
var ErrNotFound = errors.New("Credential set does not exist")

// Store is a persistent store for
type Store struct {
	autoclose    bool
	backingStore *crud.BackingStore
}

// NewCredentialStore creates a persistent store for credential sets using the specified
// backing key-blob store.
func NewCredentialStore(store crud.Store) Store {
	return Store{
		autoclose:    true,
		backingStore: crud.NewBackingStore(store),
	}
}

// List lists the names of the stored credential sets.
func (s Store) List() ([]string, error) {
	if s.autoclose {
		defer s.backingStore.Close()
	}

	return s.backingStore.List(ItemType)
}

// Save a credential set. Any previous version of the credential set is overwritten.
func (s Store) Save(cred CredentialSet) error {
	if s.autoclose {
		defer s.backingStore.Close()
	}

	bytes, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return err
	}
	return s.backingStore.Save(ItemType, cred.Name, bytes)
}

// Read loads the credential set with the given name from the store.
func (s Store) Read(name string) (CredentialSet, error) {
	if s.autoclose {
		defer s.backingStore.Close()
	}

	bytes, err := s.backingStore.Read(ItemType, name)
	if err != nil {
		if err == crud.ErrRecordDoesNotExist {
			return CredentialSet{}, ErrNotFound
		}
		return CredentialSet{}, err
	}
	credset := CredentialSet{}
	err = json.Unmarshal(bytes, &credset)
	return credset, err
}

// ReadAll retrieves all the credential sets.
func (s Store) ReadAll() ([]CredentialSet, error) {
	if s.autoclose {
		defer s.backingStore.Close()
	}

	results := make([]CredentialSet, 0)

	list, err := s.backingStore.List(ItemType)
	if err != nil {
		return results, err
	}

	autoClose := s.autoclose
	s.autoclose = false
	for _, name := range list {
		result, err := s.Read(name)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	s.autoclose = autoClose

	return results, nil
}

// Delete deletes a credential set from the store.
func (s Store) Delete(name string) error {
	if s.autoclose {
		defer s.backingStore.Close()
	}

	return s.backingStore.Delete(ItemType, name)
}
