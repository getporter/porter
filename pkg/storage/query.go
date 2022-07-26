package storage

import (
	"encoding/json"
	"strings"

	"get.porter.sh/porter/pkg/storage/plugins"
	"go.mongodb.org/mongo-driver/bson"
)

// AggregateOptions is the set of options available to the Aggregate operation on any
// storage provider.
type AggregateOptions struct {
	// Pipeline document to aggregate, filter, and shape the results.
	// See https://docs.mongodb.com/manual/reference/operator/aggregation-pipeline/
	Pipeline []bson.D
}

func (o AggregateOptions) ToPluginOptions(collection string) plugins.AggregateOptions {
	return plugins.AggregateOptions{
		Collection: collection,
		Pipeline:   o.Pipeline,
	}
}

// EnsureIndexOptions is the set of options available to the EnsureIndex operation.
type EnsureIndexOptions struct {
	// Indices to create if not found.
	Indices []Index
}

// Index on a collection.
type Index struct {
	// Collection name to which the index applies.
	Collection string

	// Keys describes the fields and their sort order.
	// Example: ["namespace", "name", "-timestamp"]
	Keys []string

	// Unique specifies if the index should enforce that the indexed fields for each document are unique.
	Unique bool
}

// Convert from a simplified sort specifier like []{"-key"}
// to a mongodb sort document like []{{Key: "key", Value: -1}}
func convertSortKeys(values []string) bson.D {
	if len(values) == 0 {
		return nil
	}

	keys := make(bson.D, len(values))
	for i, key := range values {
		sortKey := key
		sortOrder := 1
		if strings.HasPrefix(key, "-") {
			sortKey = strings.Trim(key, "-")
			sortOrder = -1
		}
		keys[i] = bson.E{Key: sortKey, Value: sortOrder}
	}
	return keys
}

func (o EnsureIndexOptions) ToPluginOptions() plugins.EnsureIndexOptions {
	opts := plugins.EnsureIndexOptions{
		Indices: make([]plugins.Index, len(o.Indices)),
	}
	for i, index := range o.Indices {
		opts.Indices[i] = plugins.Index{
			Collection: index.Collection,
			Keys:       convertSortKeys(index.Keys),
			Unique:     index.Unique,
		}
	}
	return opts
}

// CountOptions is the set of options available to the Count operation on any
// storage provider.
type CountOptions struct {
	// Query is a query filter document
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	Filter bson.M
}

func (o CountOptions) ToPluginOptions(collection string) plugins.CountOptions {
	if o.Filter == nil {
		o.Filter = map[string]interface{}{}
	}
	return plugins.CountOptions{
		Collection: collection,
		Filter:     o.Filter,
	}
}

// FindOptions is the set of options available to the StorageProtocol.Find
// operation.
type FindOptions struct {
	// Sort is a list of field names by which the results should be sorted.
	// Prefix a field with "-" to sort in reverse order.
	Sort []string

	// Skip is the number of results to skip past and exclude from the results.
	Skip int64

	// Limit is the number of results to return.
	Limit int64

	// Filter specifies a filter the results.
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	Filter bson.M

	// Select is a projection document. The entire document is returned by default.
	// See https://docs.mongodb.com/manual/tutorial/project-fields-from-query-results/
	Select bson.D
}

func (o FindOptions) ToPluginOptions(collection string) plugins.FindOptions {
	if o.Filter == nil {
		o.Filter = bson.M{}
	}
	return plugins.FindOptions{
		Collection: collection,
		Sort:       convertSortKeys(o.Sort),
		Skip:       o.Skip,
		Limit:      o.Limit,
		Select:     o.Select,
		Filter:     o.Filter,
	}
}

// GetOptions is the set of options available for the Get operation.
// Documents can be retrieved by either ID or Namespace + Name.
type GetOptions struct {
	// ID of the document to retrieve.
	ID string

	// Name of the document to retrieve.
	Name string

	// Namespace of the document to retrieve.
	Namespace string
}

// ToFindOptions converts from the convenience method Get to FindOne.
func (o GetOptions) ToFindOptions() FindOptions {
	var filter map[string]interface{}
	if o.ID != "" {
		filter = map[string]interface{}{"_id": o.ID}
	} else if o.Name != "" {
		filter = map[string]interface{}{"namespace": o.Namespace, "name": o.Name}
	}

	return FindOptions{
		Filter: filter,
	}
}

// InsertOptions is the set of options for the StorageProtocol.Insert operation.
type InsertOptions struct {
	// Documents is a set of documents to insert.
	Documents []interface{}
}

func (o InsertOptions) ToPluginOptions(collection string) (plugins.InsertOptions, error) {
	var docs []bson.M
	err := convertToRawJsonDocument(o.Documents, &docs)
	if err != nil {
		return plugins.InsertOptions{}, nil
	}

	return plugins.InsertOptions{
		Collection: collection,
		Documents:  docs,
	}, nil
}

// PatchOptions is the set of options for the StorageProtocol.Patch operation.
type PatchOptions struct {
	// Query is a query filter document
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	QueryDocument bson.M

	// Transformation is set of instructions to modify matching
	// documents.
	Transformation bson.D
}

func (o PatchOptions) ToPluginOptions(collection string) plugins.PatchOptions {
	return plugins.PatchOptions{
		Collection:     collection,
		QueryDocument:  o.QueryDocument,
		Transformation: o.Transformation,
	}
}

// RemoveOptions is the set of options for the StorageProtocol.Remove operation.
type RemoveOptions struct {
	// Filter is a query filter document
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	Filter bson.M

	// All matching documents should be removed. Defaults to false, which only
	// removes the first matching document.
	All bool

	// ID of the document to remove. This sets the Filter to an _id match using the specified value.
	ID string

	// Name of the document to remove.
	Name string

	// Namespace of the document to remove.
	Namespace string
}

func (o RemoveOptions) ToPluginOptions(collection string) plugins.RemoveOptions {
	// If a custom filter wasn't specified, update the specified document
	if o.Filter == nil {
		if o.ID != "" {
			o.Filter = map[string]interface{}{"_id": o.ID}
		} else if o.Name != "" {
			o.Filter = map[string]interface{}{"namespace": o.Namespace, "name": o.Name}
		}
	}

	return plugins.RemoveOptions{
		Collection: collection,
		Filter:     o.Filter,
		All:        o.All,
	}
}

// UpdateOptions is the set of options for the StorageProtocol.Update operation.
type UpdateOptions struct {
	// Filter is a query filter document. Defaults to filtering by the document id.
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	Filter bson.M

	// Upsert indicates that the document should be inserted if not found
	Upsert bool

	// Document is the replacement document.
	Document interface{}
}

func (o UpdateOptions) ToPluginOptions(collection string) (plugins.UpdateOptions, error) {
	// If a custom filter wasn't specified, update the specified document
	if o.Filter == nil {
		if doc, ok := o.Document.(Document); ok {
			o.Filter = doc.DefaultDocumentFilter()
		}
	}

	var doc map[string]interface{}
	err := convertToRawJsonDocument(o.Document, &doc)
	if err != nil {
		return plugins.UpdateOptions{}, nil
	}

	return plugins.UpdateOptions{
		Collection: collection,
		Filter:     o.Filter,
		Upsert:     o.Upsert,
		Document:   doc,
	}, nil
}

// Document represents a stored Porter document with
// accessor methods to make persistence more straightforward.
type Document interface {
	// DefaultDocumentFilter is the default filter to match the current document.
	DefaultDocumentFilter() map[string]interface{}
}

// converts a set of typed documents to a raw representation using maps
// by way of the type's json representation. This ensures that any
// json marshal logic is used when serializing documents to the database.
// e.g. if a document has a calculated field such as _id (which is required
// when persisting the document), that it is included in the doc sent to the database.
func convertToRawJsonDocument(in interface{}, raw interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, raw)
}

// ListOptions is the set of options available to the list operation
// on any storage provider.
type ListOptions struct {
	// Namespace in which the particular result list is defined.
	Namespace string

	// Name specifies whether the result list name contain the specified substring.
	Name string

	// Labels is used to filter result list based on a key-value pair.
	Labels map[string]string

	// Skip is the number of results to skip past and exclude from the results.
	Skip int64

	// Limit is the number of results to return.
	Limit int64
}

// ToFindOptions builds a query for a list of documents with these conditions:
// * sorted in ascending order by namespace first and then name
// * filtered by matching namespace, name contains substring, and labels contain all matches
// * skipped and limited to a certain number of result
func (o ListOptions) ToFindOptions() FindOptions {
	filter := make(map[string]interface{}, 3)
	if o.Namespace != "*" {
		filter["namespace"] = o.Namespace
	}
	if o.Name != "" {
		filter["name"] = map[string]interface{}{"$regex": o.Name}
	}
	for k, v := range o.Labels {
		filter["labels."+k] = v
	}

	return FindOptions{
		Sort:   []string{"namespace", "name"},
		Filter: filter,
		Skip:   o.Skip,
		Limit:  o.Limit,
	}
}
