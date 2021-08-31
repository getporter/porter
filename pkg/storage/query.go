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
	Pipeline interface{}
}

func (o AggregateOptions) ToPluginOptions(collection string) plugins.AggregateOptions {
	return plugins.AggregateOptions{
		Collection: collection,
		Pipeline:   o.Pipeline,
	}
}

// EnsureIndexOptions is the set of options available to the EnsureIndex operation.
type EnsureIndexOptions struct {
	Keys   []string `json:"keys"`
	Unique bool     `json:"unique"`
}

// Convert from a simplified sort specifier like []{"-key"}
// to a mongodb sort document like []{{Key: "key", Value: -1}}
func convertSortKeys(values []string) interface{} {
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
		keys[i] = bson.E{
			Key:   sortKey,
			Value: sortOrder,
		}
	}
	return keys
}

func (o EnsureIndexOptions) ToPluginOptions(collection string) plugins.EnsureIndexOptions {
	return plugins.EnsureIndexOptions{
		Collection: collection,
		Keys:       convertSortKeys(o.Keys),
		Unique:     o.Unique,
	}
}

// CountOptions is the set of options available to the Count operation on any
// storage provider.
type CountOptions struct {
	// Query is a query filter document
	// See https://docs.mongodb.com/manual/core/document/#std-label-document-query-filter
	Filter interface{}
}

func (o CountOptions) ToPluginOptions(collection string) plugins.CountOptions {
	if o.Filter == nil {
		o.Filter = bson.M{}
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
	Filter interface{}

	// Select is a projection document. The entire document is returned by default.
	// See https://docs.mongodb.com/manual/tutorial/project-fields-from-query-results/
	Select interface{}
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

// ToPluginOptions converts from the convenience method Get to FindOne.
func (o GetOptions) ToPluginOptions() FindOptions {
	var filter interface{}
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
	var docs []interface{}
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
	QueryDocument interface{}

	// Transformation is set of instructions to modify matching
	// documents.
	Transformation interface{}
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
	Filter interface{}

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
	Filter interface{}

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
	// DefaultDocumentFilter is the default filter to match the curent document.
	DefaultDocumentFilter() interface{}
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

// CreateListFiler builds a query for a list of documents by:
// * matching namespace
// * name contains substring
// * labels contains all matches
func CreateListFiler(namespace string, name string, labels map[string]string) bson.M {
	filter := make(bson.M, 3)
	if namespace != "*" {
		filter["namespace"] = namespace
	}
	if name != "" {
		filter["name"] = bson.M{"$regex": name}
	}
	for k, v := range labels {
		filter["labels."+k] = v
	}
	return filter
}
