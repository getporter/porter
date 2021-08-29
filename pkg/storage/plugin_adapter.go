package storage

import (
	"encoding/json"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/storage/plugins"
	"github.com/pkg/errors"
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
	plugin plugins.StoragePlugin
}

// NewPluginAdapter wraps the specified storage plugin.
func NewPluginAdapter(plugin plugins.StoragePlugin) PluginAdapter {
	return PluginAdapter{plugin: plugin}
}

func (a PluginAdapter) Connect() error {
	return a.plugin.Connect()
}

func (a PluginAdapter) Close() error {
	return a.plugin.Close()
}

func (a PluginAdapter) Aggregate(collection string, opts AggregateOptions, out interface{}) error {
	err := a.Connect()
	if err != nil {
		return err
	}

	rawResults, err := a.plugin.Aggregate(opts.ToPluginOptions(collection))
	if err != nil {
		return err
	}

	return a.unmarshalSlice(rawResults, out)
}

func (a PluginAdapter) EnsureIndex(collection string, opts EnsureIndexOptions) error {
	err := a.Connect()
	if err != nil {
		return err
	}

	return a.plugin.EnsureIndex(opts.ToPluginOptions(collection))
}

func (a PluginAdapter) Count(collection string, opts CountOptions) (int64, error) {
	err := a.Connect()
	if err != nil {
		return 0, err
	}

	return a.plugin.Count(opts.ToPluginOptions(collection))
}

func (a PluginAdapter) Find(collection string, opts FindOptions, out interface{}) error {
	err := a.Connect()
	if err != nil {
		return err
	}

	rawResults, err := a.plugin.Find(opts.ToPluginOptions(collection))
	if err != nil {
		return err
	}

	return a.unmarshalSlice(rawResults, out)
}

// FindOne queries a collection and returns the first result, returning
// ErrNotFound when no results are returned.
func (a PluginAdapter) FindOne(collection string, opts FindOptions, out interface{}) error {
	err := a.Connect()
	if err != nil {
		return err
	}

	rawResults, err := a.plugin.Find(opts.ToPluginOptions(collection))
	if err != nil {
		return err
	}

	if len(rawResults) == 0 {
		return ErrNotFound{Collection: collection}
	}

	err = a.unmarshal(rawResults[0], out)
	return errors.Wrapf(err, "could not unmarshal document of type %T", out)
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
		return errors.Wrap(err, "error marshaling results into a single result document")
	}

	// Unmarshal the consolidated results onto our destination output
	err = json.Unmarshal(data, out)
	return errors.Wrapf(err, "could not unmarshal slice onto type %T", out)
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
		return errors.Wrap(err, "error marshaling results into a single result document")
	}

	// Unmarshal the consolidated results onto our destination output
	err = json.Unmarshal(data, out)
	return errors.Wrapf(err, "could not unmarshal slice onto type %T", out)
}

func (a PluginAdapter) Get(collection string, opts GetOptions, out interface{}) error {
	findOpts := opts.ToPluginOptions()
	return a.FindOne(collection, findOpts, out)
}

func (a PluginAdapter) Insert(collection string, opts InsertOptions) error {
	err := a.Connect()
	if err != nil {
		return err
	}

	pluginOpts, err := opts.ToPluginOptions(collection)
	if err != nil {
		return err
	}

	return a.plugin.Insert(pluginOpts)
}

func (a PluginAdapter) Patch(collection string, opts PatchOptions) error {
	err := a.Connect()
	if err != nil {
		return err
	}

	err = a.plugin.Patch(opts.ToPluginOptions(collection))
	return a.handleError(err, collection)
}

func (a PluginAdapter) Remove(collection string, opts RemoveOptions) error {
	err := a.Connect()
	if err != nil {
		return err
	}

	err = a.plugin.Remove(opts.ToPluginOptions(collection))
	return a.handleError(err, collection)
}

func (a PluginAdapter) Update(collection string, opts UpdateOptions) error {
	err := a.Connect()
	if err != nil {
		return err
	}

	pluginOpts, err := opts.ToPluginOptions(collection)
	if err != nil {
		return err
	}

	err = a.plugin.Update(pluginOpts)
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
	case "credentials":
		docType = "Credential Set"
	case "parameters":
		docType = "Parameter Set"
	}
	return fmt.Sprintf("%s not found", docType)
}

func (e ErrNotFound) Is(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}
