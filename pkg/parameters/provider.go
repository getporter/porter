package parameters

import (
	"context"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
)

// Provider interface for managing sets of parameters.
type Provider interface {
	GetDataStore() storage.Store

	// ResolveAll parameter values in the parameter set.
	ResolveAll(ctx context.Context, params ParameterSet) (secrets.Set, error)

	// Validate the parameter set is defined properly.
	Validate(ctx context.Context, params ParameterSet) error

	InsertParameterSet(ctx context.Context, params ParameterSet) error
	ListParameterSets(ctx context.Context, namespace string, name string, labels map[string]string) ([]ParameterSet, error)
	GetParameterSet(ctx context.Context, namespace string, name string) (ParameterSet, error)
	UpdateParameterSet(ctx context.Context, params ParameterSet) error
	UpsertParameterSet(ctx context.Context, params ParameterSet) error
	RemoveParameterSet(ctx context.Context, namespace string, name string) error
}
