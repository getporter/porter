package secrets

import (
	"errors"
)

var ErrNotFound = errors.New("secret not found")

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
