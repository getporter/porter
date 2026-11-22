package storage

import (
	"context"

	"get.porter.sh/porter/pkg/storage/plugins"
)

// Provider handles high level functions over Porter's storage systems such as
// migrating data formats.
type Provider interface {
	plugins.StorageBackend

	// Close the connection to the underlying storage.
	Close() error

	// WriteSchema persists an up-to-date schema to the underlying storage.
	WriteSchema(ctx context.Context) error

	// MigrateData migrates data from an older version of Porter into the current version.
	MigrateData(ctx context.Context, opts MigrateOptions) error
}

// MigrateOptions are the set of available options to configure a storage data migration
// from an older version of Porter into the current version.
type MigrateOptions struct {
	// OldHome is the path to the PORTER_HOME directory for the previous version of porter.
	OldHome string

	// OldStorageAccount is the name of the storage account configured in MigrateOptions.OldHome
	// where records should be migrated from.
	OldStorageAccount string

	// NewNamespace is the namespace into which records should be imported.
	NewNamespace string
}
