package storage

import (
	"context"

	"get.porter.sh/porter/pkg/secrets"
)

// ParameterSetProvider interface for managing sets of parameters.
type ParameterSetProvider interface {
	// GetParameterSetWithFallback retrieves a parameter set from the specified
	// namespace, falling back to the global namespace when it isn't found there.
	GetParameterSetWithFallback(ctx context.Context, namespace string, name string) (ParameterSet, error)

	// ResolveAll parameter values in the parameter set.
	ResolveAll(ctx context.Context, params ParameterSet, keys []string) (secrets.Set, error)

	// Validate the parameter set is defined properly.
	Validate(ctx context.Context, params ParameterSet) error

	InsertParameterSet(ctx context.Context, params ParameterSet) error
	ListParameterSets(ctx context.Context, listOptions ListOptions) ([]ParameterSet, error)
	GetParameterSet(ctx context.Context, namespace string, name string) (ParameterSet, error)
	UpdateParameterSet(ctx context.Context, params ParameterSet) error
	UpsertParameterSet(ctx context.Context, params ParameterSet) error
	RemoveParameterSet(ctx context.Context, namespace string, name string) error
}
