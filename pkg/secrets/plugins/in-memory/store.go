package inmemory

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/secrets/plugins"
)

var _ plugins.SecretsProtocol = &Store{}

// Store implements an in-memory secrets store for testing.
type Store struct {
	Secrets map[string]map[string]string
}

func NewStore() *Store {
	s := &Store{
		Secrets: make(map[string]map[string]string),
	}

	return s
}

func (s *Store) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	_, ok := s.Secrets[keyName]
	if !ok {
		s.Secrets[keyName] = make(map[string]string, 1)
	}

	if keyName == "secret" {
		value, ok := s.Secrets[keyName][keyValue]
		if !ok {
			return "", errors.New("secret not found")
		}

		return value, nil
	}

	return "", fmt.Errorf("unsupported keyName %s", keyName)
}

func (s *Store) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	_, ok := s.Secrets[keyName]
	if !ok {
		s.Secrets[keyName] = make(map[string]string, 1)
	}

	s.Secrets[keyName][keyValue] = value
	return nil
}
