package secrets

import (
	"errors"

	"github.com/cnabio/cnab-go/credentials"
)

var ErrNotFound = errors.New("secret not found")

// Store interface for working with secret sources.
type Store interface {
	// Resolve a credential's value from a secret store.
	Resolve(cred credentials.Source) (string, error)
}

// HasConnect indicates that a struct must be initialized using the Connect
// method before the interface's methods are called.
type HasConnect interface {
	Connect() error
}

// HasClose indicates that a struct must be cleaned up using the Close
// method before the interface's methods are called.
type HasClose interface {
	Close() error
}
