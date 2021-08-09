package parameters

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

var _ Provider = &ParameterStore{}

const (
	CollectionParameters = "parameters"
)

// ParameterStore provides access to parameter sets by instantiating plugins that
// implement CRUD storage.
type ParameterStore struct {
	Documents storage.Store
	Secrets   secrets.Store
}

func NewParameterStore(storage storage.Store, secrets secrets.Store) *ParameterStore {
	return &ParameterStore{
		Documents: storage,
		Secrets:   secrets,
	}
}

// Initialize the backend storage with any necessary schema changes, such as indexes.
func (s ParameterStore) Initialize() error {
	// query parameters by namespace + name
	err := s.Documents.EnsureIndex(CollectionParameters, storage.EnsureIndexOptions{
		Index: mgo.Index{
			Key:        []string{"namespace", "name"},
			Unique:     true,
			Background: true,
		},
	})
	return err
}

func (s ParameterStore) ResolveAll(params ParameterSet) (secrets.Set, error) {
	resolvedParams := make(secrets.Set)
	var resolveErrors error

	for _, param := range params.Parameters {
		value, err := s.Secrets.Resolve(param.Source.Key, param.Source.Value)
		if err != nil {
			resolveErrors = multierror.Append(resolveErrors, errors.Wrapf(err, "unable to resolve parameter %s.%s from %s %s", params.Name, param.Name, param.Source.Key, param.Source.Value))
		}

		resolvedParams[param.Name] = value
	}

	return resolvedParams, resolveErrors
}

func (s ParameterStore) Validate(params ParameterSet) error {
	validSources := []string{secrets.SourceSecret, host.SourceValue, host.SourceEnv, host.SourcePath, host.SourceCommand}
	var errors error

	for _, cs := range params.Parameters {
		valid := false
		for _, validSource := range validSources {
			if cs.Source.Key == validSource {
				valid = true
				break
			}
		}
		if valid == false {
			errors = multierror.Append(errors, fmt.Errorf(
				"%s is not a valid source. Valid sources are: %s",
				cs.Source.Key,
				strings.Join(validSources, ", "),
			))
		}
	}

	return errors
}

func (s ParameterStore) InsertParameterSet(cset ParameterSet) error {
	opts := storage.InsertOptions{
		Documents: []interface{}{cset},
	}
	return s.Documents.Insert(CollectionParameters, opts)
}

func (s ParameterStore) ListParameterSets(namespace string) ([]ParameterSet, error) {
	var out []ParameterSet
	opts := storage.FindOptions{}
	if namespace != "*" {
		opts.Filter = bson.M{"namespace": namespace}
	}
	err := s.Documents.Find(CollectionParameters, opts, &out)
	return out, err
}

func (s ParameterStore) GetParameterSet(namespace string, name string) (ParameterSet, error) {
	var out ParameterSet
	opts := storage.FindOptions{
		Filter: map[string]interface{}{
			"namespace": namespace,
			"name":      name,
		},
	}
	err := s.Documents.FindOne(CollectionParameters, opts, &out)
	return out, err
}

func (s ParameterStore) UpdateParameterSet(params ParameterSet) error {
	opts := storage.UpdateOptions{
		Document: params,
	}
	return s.Documents.Update(CollectionParameters, opts)
}

func (s ParameterStore) UpsertParameterSet(params ParameterSet) error {
	opts := storage.UpdateOptions{
		Document: params,
		Upsert:   true,
	}
	return s.Documents.Update(CollectionParameters, opts)
}

func (s ParameterStore) RemoveParameterSet(namespace string, name string) error {
	opts := storage.RemoveOptions{
		Namespace: namespace,
		Name:      name,
	}
	return s.Documents.Remove(CollectionParameters, opts)
}
