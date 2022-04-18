package plugins

import (
	"context"

	"get.porter.sh/porter/pkg/plugins"
	"go.mongodb.org/mongo-driver/bson"
)

// PluginInterface for the data storage. This first part of the
// three-part plugin key is only seen/used by the plugins when the host is
// communicating with the plugin and is not exposed to users.
const PluginInterface = "storage"

// StoragePlugin is the interface used to wrap a storage plugin.
// It is not meant to be used directly.
type StoragePlugin interface {
	plugins.Plugin

	// EnsureIndex makes sure that the specified index exists as specified.
	// If it does exist with a different definition, the index is recreated.
	EnsureIndex(ctx context.Context, opts EnsureIndexOptions) error

	// Aggregate executes a pipeline and returns the results.
	Aggregate(ctx context.Context, opts AggregateOptions) ([]bson.Raw, error)

	// Count the number of results that match an optional query.
	// When the query is omitted, the entire collection is counted.
	Count(ctx context.Context, opts CountOptions) (int64, error)

	// Find queries a collection, optionally projecting a subset of fields, and
	// then returns the results as a list of bson documents.
	Find(ctx context.Context, opts FindOptions) ([]bson.Raw, error)

	// Insert a set of documents into a collection.
	Insert(ctx context.Context, opts InsertOptions) error

	// Patch applies a transformation to matching documents.
	Patch(ctx context.Context, opts PatchOptions) error

	// Remove matching documents from a collection.
	Remove(ctx context.Context, opts RemoveOptions) error

	// Update matching documents with the specified replacement document.
	Update(ctx context.Context, opts UpdateOptions) error
}
