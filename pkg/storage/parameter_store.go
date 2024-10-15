package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"

	"get.porter.sh/porter/pkg/secrets"
	hostSecrets "get.porter.sh/porter/pkg/secrets/plugins/host"
	"get.porter.sh/porter/pkg/tracing"
)

var _ ParameterSetProvider = &ParameterStore{}

const (
	CollectionParameters = "parameters"
)

// ParameterStore provides access to parameter sets by instantiating plugins that
// implement CRUD storage.
type ParameterStore struct {
	Documents   Store
	Secrets     secrets.Store
	HostSecrets hostSecrets.Store
}

func NewParameterStore(storage Store, secrets secrets.Store) *ParameterStore {
	return &ParameterStore{
		Documents:   storage,
		Secrets:     secrets,
		HostSecrets: hostSecrets.NewStore(),
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
	return resolveAll(ctx, params.Parameters, s.HostSecrets, s.Secrets, params.Name, "parameter")
}

func resolveAll(
	ctx context.Context,
	list secrets.StrategyList,
	hostSecrets hostSecrets.Store,
	secretsStore secrets.Store,
	name string,
	kind string,
) (secrets.Set, error) {
	resolvedParams := make(secrets.Set)
	var resolveErrors error

	for _, srcMap := range list {
		var value string
		var err error
		if isHandledByHostPlugin(srcMap.Source.Strategy) {
			value, err = hostSecrets.Resolve(ctx, srcMap.Source.Strategy, srcMap.Source.Hint)
		} else {
			value, err = secretsStore.Resolve(ctx, srcMap.Source.Strategy, srcMap.Source.Hint)
		}
		if err != nil {
			resolveErrors = multierror.Append(resolveErrors, fmt.Errorf("unable to resolve %s %s.%s from %s %s: %w", kind, name, srcMap.Name, srcMap.Source.Strategy, srcMap.Source.Hint, err))
		}

		resolvedParams[srcMap.Name] = value
	}

	return resolvedParams, resolveErrors
}

func (s ParameterStore) Validate(ctx context.Context, params ParameterSet) error {
	return validate(params.Parameters)
}

func validate(list secrets.StrategyList) error {
	validSources := []string{secrets.SourceSecret, host.SourceValue, host.SourceEnv, host.SourcePath, host.SourceCommand}
	var errors error

	for _, cs := range list {
		valid := false
		for _, validSource := range validSources {
			if cs.Source.Strategy == validSource {
				valid = true
				break
			}
		}
		if !valid {
			errors = multierror.Append(errors, fmt.Errorf(
				"%s is not a valid source. Valid sources are: %s",
				cs.Source.Strategy,
				strings.Join(validSources, ", "),
			))
		}
	}

	return errors
}

func (s ParameterStore) InsertParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = DefaultParameterSetSchemaVersion
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
func (s ParameterStore) FindParameterSet(ctx context.Context, namespace string, name string) (ParameterSet, error) {
	var out ParameterSet
	query := FindOptions{
		Sort: []string{"-namespace"},
		Filter: bson.M{
			"name": name,
			"$or": []bson.M{
				{"namespace": ""},
				{"namespace": namespace},
			},
		},
	}
	err := s.Documents.FindOne(ctx, CollectionCredentials, query, &out)
	return out, err
}

func (s ParameterStore) UpdateParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = DefaultParameterSetSchemaVersion
	opts := UpdateOptions{
		Document: params,
	}
	return s.Documents.Update(ctx, CollectionParameters, opts)
}

func (s ParameterStore) UpsertParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = DefaultParameterSetSchemaVersion
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
