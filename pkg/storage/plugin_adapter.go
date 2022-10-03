package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"get.porter.sh/porter/pkg/storage/plugins"
	"go.mongodb.org/mongo-driver/bson"
)

var _ Store = PluginAdapter{}

// PluginAdapter converts between the low-level plugin.StorageProtocol which
// operates on bson documents, and the document types stored by Porter which are
// marshaled using json.
//
// Specifically it handles converting from bson.Raw to the type specified by
// ResultType on plugin.ResultOptions so that you can just cast the result to
// the specified type safely.
type PluginAdapter struct {
	plugin plugins.StorageProtocol
}

// NewPluginAdapter wraps the specified storage plugin.
func NewPluginAdapter(plugin plugins.StorageProtocol) PluginAdapter {
	return PluginAdapter{
		plugin: plugin,
	}
}

func (a PluginAdapter) Close() error {
	if closer, ok := a.plugin.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (a PluginAdapter) Aggregate(ctx context.Context, collection string, opts AggregateOptions, out interface{}) error {
	rawResults, err := a.plugin.Aggregate(ctx, opts.ToPluginOptions(collection))
	if err != nil {
		return err
	}

	return a.unmarshalSlice(rawResults, out)
}

func (a PluginAdapter) EnsureIndex(ctx context.Context, opts EnsureIndexOptions) error {
	return a.plugin.EnsureIndex(ctx, opts.ToPluginOptions())
}

func (a PluginAdapter) Count(ctx context.Context, collection string, opts CountOptions) (int64, error) {
	return a.plugin.Count(ctx, opts.ToPluginOptions(collection))
}

func (a PluginAdapter) Find(ctx context.Context, collection string, opts FindOptions, out interface{}) error {
	rawResults, err := a.plugin.Find(ctx, opts.ToPluginOptions(collection))
	if err != nil {
		return a.handleError(err, collection)

	}

	return a.unmarshalSlice(rawResults, out)
}

// FindOne queries a collection and returns the first result, returning
// ErrNotFound when no results are returned.
func (a PluginAdapter) FindOne(ctx context.Context, collection string, opts FindOptions, out interface{}) error {
	rawResults, err := a.plugin.Find(ctx, opts.ToPluginOptions(collection))
	if err != nil {
		return a.handleError(err, collection)
	}

	if len(rawResults) == 0 {
		notFoundErr := ErrNotFound{Collection: collection}
		if name, ok := opts.Filter["name"]; ok {
			notFoundErr.Item = fmt.Sprint(name)
		}
		return notFoundErr
	}

	err = a.unmarshal(rawResults[0], out)
	if err != nil {
		return fmt.Errorf("could not unmarshal document of type %T: %w", out, err)
	}

	return nil
}

// unmarshalSlice unpacks a slice of bson documents onto the specified type slice (out)
// by going through a temporary representation of the document as json so that we
// use the json marshal logic defined on the struct, e.g. if fields have different
// names defined with json tags.
func (a PluginAdapter) unmarshalSlice(bsonResults []bson.Raw, out interface{}) error {
	// We want to go from []bson.Raw -> []bson.M -> json -> out (typed struct)

	// Populate a single document with all the results in an intermediate
	// format of map[string]interface
	tmpResults := make([]bson.M, len(bsonResults))
	for i, bsonResult := range bsonResults {
		var result bson.M
		err := bson.Unmarshal(bsonResult, &result)
		if err != nil {
			return err
		}
		tmpResults[i] = result
	}

	// Marshal the consolidated document to json
	data, err := json.Marshal(tmpResults)
	if err != nil {
		return fmt.Errorf("error marshaling results into a single result document: %w", err)
	}

	// Unmarshal the consolidated results onto our destination output
	err = json.Unmarshal(data, out)
	if err != nil {
		return fmt.Errorf("could not unmarshal slice onto type %T: %w", out, err)
	}

	return nil
}

// unmarshalSlice a bson document onto the specified typed output
// by going through a temporary representation of the document as json so that we
// use the json marshal logic defined on the struct, e.g. if fields have different
// names defined with json tags.
func (a PluginAdapter) unmarshal(bsonResult bson.Raw, out interface{}) error {
	// We want to go from bson.Raw -> bson.M -> json -> out (typed struct)

	var tmpResult bson.M
	err := bson.Unmarshal(bsonResult, &tmpResult)
	if err != nil {
		return err
	}

	// Marshal the consolidated document to json
	data, err := json.Marshal(tmpResult)
	if err != nil {
		return fmt.Errorf("error marshaling results into a single result document: %w", err)
	}

	// Unmarshal the consolidated results onto our destination output
	err = json.Unmarshal(data, out)
	if err != nil {
		return fmt.Errorf("could not unmarshal slice onto type %T: %w", out, err)
	}

	return nil
}

func (a PluginAdapter) Get(ctx context.Context, collection string, opts GetOptions, out interface{}) error {
	findOpts := opts.ToFindOptions()
	err := a.FindOne(ctx, collection, findOpts, out)
	return a.handleError(err, collection)
}

func (a PluginAdapter) Insert(ctx context.Context, collection string, opts InsertOptions) error {
	pluginOpts, err := opts.ToPluginOptions(collection)
	if err != nil {
		return err
	}

	err = a.plugin.Insert(ctx, pluginOpts)
	return a.handleError(err, collection)
}

func (a PluginAdapter) Patch(ctx context.Context, collection string, opts PatchOptions) error {
	err := a.plugin.Patch(ctx, opts.ToPluginOptions(collection))
	return a.handleError(err, collection)
}

func (a PluginAdapter) Remove(ctx context.Context, collection string, opts RemoveOptions) error {
	err := a.plugin.Remove(ctx, opts.ToPluginOptions(collection))
	return a.handleError(err, collection)
}

func (a PluginAdapter) Update(ctx context.Context, collection string, opts UpdateOptions) error {
	pluginOpts, err := opts.ToPluginOptions(collection)
	if err != nil {
		return err
	}
	err = a.plugin.Update(ctx, pluginOpts)
	return a.handleError(err, collection)
}

// handleError unwraps errors returned from a plugin (which due to the round trip
// through the plugin framework means the original typed error may not be the right type anymore
// and turns it back into a well known error such as NotFound.
func (a PluginAdapter) handleError(err error, collection string) error {
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
		return ErrNotFound{Collection: collection}
	}
	return err
}

// ErrNotFound indicates that the requested document was not found.
// You can test for this error using errors.Is(err, storage.ErrNotFound{})
type ErrNotFound struct {
	Collection string
	Item       string
}

func (e ErrNotFound) Error() string {
	var docType string
	switch e.Collection {
	case "installations":
		docType = "Installation"
	case "runs":
		docType = "Run"
	case "results":
		docType = "Result"
	case "output":
		docType = "Output"
	case "credentials", "parameters":
		if len(e.Item) > 0 {
			docType = e.Item
		}
	}

	return fmt.Sprintf("%s not found", docType)
}

func (e ErrNotFound) Is(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}
