package storage

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

var _ ParameterSetProvider = &ParameterStore{}

const (
	CollectionParameters = "parameters"
)

// ParameterStore provides access to parameter sets by instantiating plugins that
// implement CRUD storage.
type ParameterStore struct {
	Documents Store
	Secrets   secrets.Store
}

func NewParameterStore(storage Store, secrets secrets.Store) *ParameterStore {
	return &ParameterStore{
		Documents: storage,
		Secrets:   secrets,
	}
}

// EnsureParameterIndices creates indices on the parameters collection.
func EnsureParameterIndices(ctx context.Context, store Store) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debug("Initializing parameter collection indices")
	indices := EnsureIndexOptions{
		Indices: []Index{
			// query parameters by namespace + name
			{Collection: CollectionParameters, Keys: []string{"namespace", "name"}, Unique: true},
		},
	}
	err := store.EnsureIndex(ctx, indices)
	return span.Error(err)
}

func (s ParameterStore) GetDataStore() Store {
	return s.Documents
}

func (s ParameterStore) ResolveAll(ctx context.Context, params ParameterSet) (secrets.Set, error) {
	resolvedParams := make(secrets.Set)
	var resolveErrors error

	for _, param := range params.Parameters {
		value, err := s.Secrets.Resolve(ctx, param.Source.Key, param.Source.Value)
		if err != nil {
			resolveErrors = multierror.Append(resolveErrors, errors.Wrapf(err, "unable to resolve parameter %s.%s from %s %s", params.Name, param.Name, param.Source.Key, param.Source.Value))
		}

		resolvedParams[param.Name] = value
	}

	return resolvedParams, resolveErrors
}

func (s ParameterStore) Validate(ctx context.Context, params ParameterSet) error {
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

func (s ParameterStore) InsertParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = ParameterSetSchemaVersion
	opts := InsertOptions{
		Documents: []interface{}{params},
	}
	return s.Documents.Insert(ctx, CollectionParameters, opts)
}

func (s ParameterStore) ListParameterSets(ctx context.Context, listOptions ListOptions) ([]ParameterSet, error) {
	var out []ParameterSet
	err := s.Documents.Find(ctx, CollectionParameters, listOptions.ToFindOptions(), &out)
	return out, err
}

func (s ParameterStore) GetParameterSet(ctx context.Context, namespace string, name string) (ParameterSet, error) {
	var out ParameterSet
	opts := FindOptions{
		Filter: map[string]interface{}{
			"namespace": namespace,
			"name":      name,
		},
	}
	err := s.Documents.FindOne(ctx, CollectionParameters, opts, &out)
	return out, err
}

func (s ParameterStore) UpdateParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = ParameterSetSchemaVersion
	opts := UpdateOptions{
		Document: params,
	}
	return s.Documents.Update(ctx, CollectionParameters, opts)
}

func (s ParameterStore) UpsertParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = ParameterSetSchemaVersion
	opts := UpdateOptions{
		Document: params,
		Upsert:   true,
	}
	return s.Documents.Update(ctx, CollectionParameters, opts)
}

func (s ParameterStore) RemoveParameterSet(ctx context.Context, namespace string, name string) error {
	opts := RemoveOptions{
		Namespace: namespace,
		Name:      name,
	}
	return s.Documents.Remove(ctx, CollectionParameters, opts)
}
