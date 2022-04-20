package storage

import "context"

// Store is an interface for managing Porter documents.
type Store interface {
	// Connect establishes a connection to the storage backend.
	// Safe to call multiple times, the existing connection is reused.
	Connect(ctx context.Context) error

	// Close the connection to the storage backend.
	Close(ctx context.Context) error

	// Aggregate executes a pipeline and returns the results.
	Aggregate(ctx context.Context, collection string, opts AggregateOptions, out interface{}) error

	// Count the number of results that match an optional query.
	// When the query is omitted, the entire collection is counted.
	Count(ctx context.Context, collection string, opts CountOptions) (int64, error)

	// EnsureIndex makes sure that the specified index exists as specified.
	// If it does exist with a different definition, the index is recreated.
	EnsureIndex(ctx context.Context, opts EnsureIndexOptions) error

	// Find queries a collection, optionally projecting a subset of fields, into
	// the specified out value.
	Find(ctx context.Context, collection string, opts FindOptions, out interface{}) error

	// FindOne queries a collection, optionally projecting a subset of fields,
	// returning the first result onto the specified out value.
	// Returns ErrNotFound when the query yields no results.
	FindOne(ctx context.Context, collection string, opts FindOptions, out interface{}) error

	// Get the document specified by its ID into the specified out value.
	// This is a convenience wrapper around FindOne for situations where you
	// are retrieving a well-known single document.
	Get(ctx context.Context, collection string, opts GetOptions, out interface{}) error

	// Insert a set of documents into a collection.
	Insert(ctx context.Context, collection string, opts InsertOptions) error

	// Patch applies a transformation to matching documents.
	Patch(ctx context.Context, collection string, opts PatchOptions) error

	// Remove matching documents from a collection.
	Remove(ctx context.Context, collection string, opts RemoveOptions) error

	// Update matching documents with the specified replacement document.
	Update(ctx context.Context, collection string, opts UpdateOptions) error
}
