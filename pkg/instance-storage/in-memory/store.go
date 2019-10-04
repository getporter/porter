package inmemory

import (
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/pkg/errors"
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
	claim, ok := s.claims[name]
	if !ok {
		return nil, errors.Errorf("claim %s not found", name)
	}
	return claim, nil
}

func (s *Store) Delete(name string) error {
	delete(s.claims, name)
	return nil
}
