package inmemory

import (
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
)

var _ crud.Store = &Store{}

type Store struct {
	claims map[string][]byte
}

func NewStore() *Store {
	s := &Store{
		claims: make(map[string][]byte),
	}

	return s
}

func (s *Store) List() ([]string, error) {
	names := make([]string, 0, len(s.claims))

	for name := range s.claims {
		names = append(names, name)
	}

	return names, nil
}

func (s *Store) Store(name string, data []byte) error {
	s.claims[name] = data
	return nil
}

func (s *Store) Read(name string) ([]byte, error) {
	c, ok := s.claims[name]
	if !ok {
		return nil, claim.ErrClaimNotFound
	}
	return c, nil
}

func (s *Store) Delete(name string) error {
	delete(s.claims, name)
	return nil
}
