package inmemory

import (
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/credentials"
)

var _ secrets.Store = &Store{}

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

func (s *Store) Resolve(cred credentials.Source) (string, error) {
	value, ok := s.Secrets[cred.Value]
	if ok {
		return value, nil
	}

	return "", secrets.ErrNotFound
}
