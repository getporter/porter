package storage

import (
	"context"
)

// Provider handles high level functions over Porter's storage systems such as
// migrating data formats.
type Provider interface {
	Store

	// WriteSchema persists an up-to-date schema to the underlying storage.
	WriteSchema(ctx context.Context) error

	// Migrate executes a migration on any/all of Porter's storage sub-systems.
	Migrate(ctx context.Context, dataCleaner *Sanitizer, opts MigrateOptions) error
}

// MigrateOptions are the set of available options to configure a storage data migration
// from an older version of Porter into the current version.
type MigrateOptions struct {
	// OldHome is the path to the PORTER_HOME directory for the previous version of porter.
	OldHome string

	// SourceAccount is the name of the storage account configured in MigrateOptions.OldHome
	// where records should be migrated from.
	SourceAccount string

	// DestinationNamespace is the namespace into which records should be imported.
	DestinationNamespace string
}
