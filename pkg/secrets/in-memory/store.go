package inmemory

import (
	"get.porter.sh/porter/pkg/secrets"
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
)

var _ cnabsecrets.Store = &Store{}

// Store implements an in-memory secrets store for testing.
type Store struct {
	Secrets map[string]string
}

func NewStore() *Store {
	s := &Store{
		Secrets: make(map[string]string),
	}

	return s
}

func (s *Store) Resolve(keyName string, keyValue string) (string, error) {
	value, ok := s.Secrets[keyValue]
	if ok {
		return value, nil
	}

	return "", secrets.ErrNotFound
}
