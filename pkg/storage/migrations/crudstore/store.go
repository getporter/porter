package crudstore

// Store is an interface for interacting with legacy crudstore plugins to migrate porter's data
type Store interface {
	// List the names of the items of the optional type and group.
	List(itemType string, group string) ([]string, error)

	// Read the data for a named item of the optional type.
	Read(itemType string, name string) ([]byte, error)
}
