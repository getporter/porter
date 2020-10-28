package secrets

import (
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/cnabio/cnab-go/utils/crud"
)

var _ cnabsecrets.Store = &SecretStore{}

// SecretStore wraps a source of secrets, that may have Connect/Close methods.
type SecretStore struct {
	AutoClose    bool
	closed       bool
	backingStore cnabsecrets.Store
}

func NewSecretStore(store cnabsecrets.Store) *SecretStore {
	return &SecretStore{
		AutoClose:    true,
		closed:       true,
		backingStore: store,
	}
}

func (s SecretStore) Connect() error {
	if !s.closed {
		return nil
	}

	if connectable, ok := s.backingStore.(crud.HasConnect); ok {
		s.closed = false
		return connectable.Connect()
	}

	return nil
}

func (s SecretStore) Close() error {
	if closable, ok := s.backingStore.(crud.HasClose); ok {
		s.closed = true
		return closable.Close()
	}
	return nil
}

func (s SecretStore) Resolve(keyName string, keyValue string) (string, error) {
	err := s.Connect()
	if err != nil {
		return "", err
	}

	if s.AutoClose {
		defer s.Close()
	}

	return s.backingStore.Resolve(keyName, keyValue)
}
