package inmemory

import (
	"github.com/cnabio/cnab-go/utils/crud"
)

var _ crud.Store = &Store{}

type Store struct {
	data map[string]map[string][]byte
}

func NewStore() *Store {
	s := &Store{
		data: make(map[string]map[string][]byte),
	}

	return s
}

func (s *Store) List(itemType string) ([]string, error) {
	names := make([]string, 0, len(s.data[itemType]))

	for name := range s.data[itemType] {
		names = append(names, name)
	}

	return names, nil
}

func (s *Store) Save(itemType string, name string, data []byte) error {
	if _, ok := s.data[itemType]; !ok {
		s.data[itemType] = make(map[string][]byte, 1)
	}
	s.data[itemType][name] = data
	return nil
}

func (s *Store) Read(itemType string, name string) ([]byte, error) {
	if _, ok := s.data[itemType]; !ok {
		return nil, crud.ErrRecordDoesNotExist
	}
	c, ok := s.data[itemType][name]
	if !ok {
		return nil, crud.ErrRecordDoesNotExist
	}
	return c, nil
}

func (s *Store) Delete(itemType string, name string) error {
	if _, ok := s.data[itemType]; !ok {
		return nil
	}
	delete(s.data[itemType], name)
	return nil
}
