package secrets

import (
	"github.com/cnabio/cnab-go/credentials"
)

var _ Store = &SecretStore{}

// SecretStore wraps another store that may have Connect/Close methods that need to be called.
type SecretStore struct {
	backingStore Store
}

func NewSecretStore(store Store) *SecretStore {
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

func (s SecretStore) Resolve(cred credentials.Source) (string, error) {
	err := s.Connect()
	if err != nil {
		return "", err
	}

	return s.backingStore.Resolve(cred)
}
