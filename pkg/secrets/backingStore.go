package secrets

import (
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
)

var _ cnabsecrets.Store = &SecretStore{}

// SecretStore wraps a source of secrets, that may have Connect/Close methods.
type SecretStore struct {
	backingStore cnabsecrets.Store
}

func NewSecretStore(store cnabsecrets.Store) *SecretStore {
	return &SecretStore{
		backingStore: store,
	}
}

func (s SecretStore) Connect() error {
	if connectable, ok := s.backingStore.(HasConnect); ok {
		return connectable.Connect()
	}

	return nil
}

func (s SecretStore) Close() error {
	var err error
	if closable, ok := s.backingStore.(HasClose); ok {
		err = closable.Close()
	}
	return err
}

func (s SecretStore) Resolve(keyName string, keyValue string) (string, error) {
	err := s.Connect()
	if err != nil {
		return "", err
	}

	return s.backingStore.Resolve(keyName, keyValue)
}
