package storage

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	hostSecrets "get.porter.sh/porter/pkg/secrets/plugins/host"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/hashicorp/go-multierror"
)

var _ ParameterSetProvider = &ParameterStore{}

const (
	CollectionParameters = "parameters"
)

// ParameterStore provides access to parameter sets by instantiating plugins that
// implement CRUD storage.
type ParameterStore struct {
	Documents   BackendClient
	Secrets     secrets.Store
	HostSecrets hostSecrets.Store
}

func NewParameterStore(storage BackendClient, secrets secrets.Store) *ParameterStore {
	return &ParameterStore{
		Documents:   storage,
		Secrets:     secrets,
		HostSecrets: hostSecrets.NewStore(),
	}
}

func (s ParameterStore) ResolveAll(ctx context.Context, params ParameterSet, keys []string) (secrets.Set, error) {
	resolvedParams := make(secrets.Set)
	var resolveErrors error

	for _, param := range params.Parameters {
		if !slices.Contains(keys, param.Name) {
			continue
		}

		var value string
		var err error
		if isHandledByHostPlugin(param.Source.Strategy) {
			value, err = s.HostSecrets.Resolve(ctx, param.Source.Strategy, param.Source.Hint)
		} else {
			value, err = s.Secrets.Resolve(ctx, param.Source.Strategy, param.Source.Hint)
		}
		if err != nil {
			resolveErrors = multierror.Append(resolveErrors, fmt.Errorf("unable to resolve parameter %s.%s from %s %s: %w", params.Name, param.Name, param.Source.Strategy, param.Source.Hint, err))
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
	return s.Documents.put(ctx, protov4.PutRequest_INSERT, params)
}

func (s ParameterStore) ListParameterSets(ctx context.Context, listOptions ListOptions) ([]ParameterSet, error) {
	return listDocuments[ParameterSet](ctx, s.Documents, listOptions.toListRequest(protov4.Kind_KIND_PARAMETER_SET))
}

func (s ParameterStore) GetParameterSet(ctx context.Context, namespace string, name string) (ParameterSet, error) {
	return getDocument[ParameterSet](ctx, s.Documents, &protov4.GetRequest{
		Kind:      protov4.Kind_KIND_PARAMETER_SET,
		Namespace: namespace,
		Name:      name,
	})
}

// GetParameterSetWithFallback retrieves the parameter set from the specified
// namespace, falling back to the global namespace when it isn't found there.
func (s ParameterStore) GetParameterSetWithFallback(ctx context.Context, namespace string, name string) (ParameterSet, error) {
	return getDocument[ParameterSet](ctx, s.Documents, &protov4.GetRequest{
		Kind:             protov4.Kind_KIND_PARAMETER_SET,
		Namespace:        namespace,
		Name:             name,
		FallbackToGlobal: true,
	})
}

func (s ParameterStore) UpdateParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = DefaultParameterSetSchemaVersion
	return s.Documents.put(ctx, protov4.PutRequest_UPDATE, params)
}

func (s ParameterStore) UpsertParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = DefaultParameterSetSchemaVersion
	return s.Documents.put(ctx, protov4.PutRequest_UPSERT, params)
}

func (s ParameterStore) RemoveParameterSet(ctx context.Context, namespace string, name string) error {
	_, err := s.Documents.backend.Remove(ctx, &protov4.RemoveRequest{
		Kind:      protov4.Kind_KIND_PARAMETER_SET,
		Namespace: namespace,
		Name:      name,
	})
	return err
}
