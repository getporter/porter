package storage

import "context"

// Provider handles high level functions over Porter's storage systems such as
// migrating data formats.
type Provider interface {
	Store

	// WriteSchema persists an up-to-date schema to the underlying storage.
	WriteSchema(ctx context.Context) error

	// Migrate executes a migration on any/all of Porter's storage sub-systems.
	Migrate(ctx context.Context) (string, error)
}
