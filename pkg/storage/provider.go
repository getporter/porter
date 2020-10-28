package storage

// StorageProvider handles high level functions over Porter's storage systems such as
// migrating data formats.
type StorageProvider interface {
	// Migrate executes a migration on any/all of Porter's storage sub-systems.
	Migrate() (string, error)
}
