package inmemory

import (
	"fmt"
	"strconv"

	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/pkg/errors"
)

var _ cnabsecrets.Store = &Store{}

const (
	// MockStoreType is a bucket where Connect and Close calls are recorded.
	MockStoreType = "mock-store"

	// ConnectCount records the number of times Connect has been called.
	ConnectCount = "connect-count"

	// CloseCount records the number of times Close has been called.
	CloseCount = "close-count"
)

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

func (s *Store) Connect() error {
	// Keep track of Connect calls for test asserts later
	count, err := s.GetConnectCount()
	if err != nil {
		return err
	}

	s.Secrets[MockStoreType][ConnectCount] = strconv.Itoa(count + 1)

	return nil
}

func (s *Store) Close() error {
	// Keep track of Close calls for test asserts later
	count, err := s.GetCloseCount()
	if err != nil {
		return err
	}

	s.Secrets[MockStoreType][CloseCount] = strconv.Itoa(count + 1)
	return nil
}

func (s *Store) Resolve(keyName string, keyValue string) (string, error) {
	_, ok := s.Secrets[keyName]
	if !ok {
		s.Secrets[keyName] = make(map[string]string, 1)
	}

	value, ok := s.Secrets[keyName][keyValue]
	if !ok {
		return "", errors.New("secret not found")
	}

	return value, nil
}

// GetConnectCount is for tests to safely read the Connect call count
// without accidentally triggering it by using Read.
func (s *Store) GetConnectCount() (int, error) {
	_, ok := s.Secrets[MockStoreType]
	if !ok {
		s.Secrets[MockStoreType] = make(map[string]string, 1)
	}

	countB, ok := s.Secrets[MockStoreType][ConnectCount]
	if !ok {
		countB = "0"
	}

	count, err := strconv.Atoi(countB)
	if err != nil {
		return 0, fmt.Errorf("could not convert connect-count %s to int: %v", countB, err)
	}

	return count, nil
}

// GetCloseCount is for tests to safely read the Close call count
// without accidentally triggering it by using Read.
func (s *Store) GetCloseCount() (int, error) {
	_, ok := s.Secrets[MockStoreType]
	if !ok {
		s.Secrets[MockStoreType] = make(map[string]string, 1)
	}

	countB, ok := s.Secrets[MockStoreType][CloseCount]
	if !ok {
		countB = "0"
	}

	count, err := strconv.Atoi(countB)
	if err != nil {
		return 0, fmt.Errorf("could not convert close-count %s to int: %v", countB, err)
	}

	return count, nil
}
