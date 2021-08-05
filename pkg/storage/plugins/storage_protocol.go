package plugins

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// StorageProtocol is the interface that storage plugins must implement.
// This defines the protocol used to communicate with storage plugins.
type StorageProtocol interface {
	// EnsureIndex makes sure that the specified index exists as specified.
	// If it does exist with a different definition, the index is recreated.
	EnsureIndex(opts EnsureIndexOptions) error

	// Aggregate executes a pipeline and returns the results.
	Aggregate(opts AggregateOptions) ([]bson.Raw, error)

	// Count the number of results that match an optional query.
	// When the query is omitted, the entire collection is counted.
	Count(opts CountOptions) (int, error)

	// Find queries a collection, optionally projecting a subset of fields, and
	// then returns the results as a list of bson documents.
	Find(opts FindOptions) ([]bson.Raw, error)

	// Insert a set of documents into a collection.
	Insert(opts InsertOptions) error

	// Patch applies a transformation to matching documents.
	Patch(opts PatchOptions) error

	// Remove matching documents from a collection.
	Remove(opts RemoveOptions) error

	// Update matching documents with the specified replacement document.
	Update(opts UpdateOptions) error
}

// EnsureIndexOptions is the set of options available to the
// StorageProtocol.EnsureIndex operation.
type EnsureIndexOptions struct {
	Collection string `bson:"collection"`
	mgo.Index  `bson:",inline"`
}

// AggregateOptions is the set of options available to the
// StorageProtocol.Aggregate operation.
type AggregateOptions struct {
	// Collection to query.
	Collection string

	// Pipeline document to aggregate, filter, and shape the results.
	// See https://docs.mongodb.com/manual/reference/operator/aggregation-pipeline/
	Pipeline interface{}
}

// CountOptions is the set of options available to the StorageProtocol.Count
// operation.
type CountOptions struct {
	// Collection to query.
	Collection string

	// Query is a query filter document
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	Filter interface{}
}

// FindOptions is the set of options available to the StorageProtocol.Find
// operation.
type FindOptions struct {
	// Collection to query.
	Collection string

	// Sort is a list of field names by which the results should be sorted.
	// Prefix a field with "-" to sort in reverse order.
	Sort []string

	// Skip is the number of results to skip past and exclude from the results.
	Skip int

	// Limit is the number of results to return.
	Limit int

	// Select is a projection document
	// See https://docs.mongodb.com/manual/tutorial/project-fields-from-query-results/
	Select interface{}

	// Filter specifies how to filter the results.
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	Filter interface{}

	// Group specifies how to group the results.
	// See https://docs.mongodb.com/manual/reference/operator/aggregation/group/
	Group interface{}
}

// InsertOptions is the set of options for the StorageProtocol.Insert operation.
type InsertOptions struct {
	// Collection to query.
	Collection string

	// Documents is a set of documents to insert.
	Documents []interface{}
}

// PatchOptions is the set of options for the StorageProtocol.Patch operation.
type PatchOptions struct {
	// Collection to query.
	Collection string

	// Query is a query filter document
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	QueryDocument interface{}

	// Transformation is set of instructions to modify matching
	// documents.
	Transformation interface{}
}

// RemoveOptions is the set of options for the StorageProtocol.Remove operation.
type RemoveOptions struct {
	// Collection to query.
	Collection string

	// Filter is a query filter document
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	Filter interface{}

	// All matching documents should be removed. Defaults to false, which only
	// removes the first matching document.
	All bool
}

// UpdateOptions is the set of options for the StorageProtocol.Update operation.
type UpdateOptions struct {
	// Collection to query.
	Collection string

	// Filter is a query filter document
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	Filter interface{}

	// Upsert indicates that the document should be inserted if not found
	Upsert bool

	// Document is the replacement document.
	Document interface{}
}
