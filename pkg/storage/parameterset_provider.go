package storage

import (
	"context"

	"get.porter.sh/porter/pkg/secrets"
)

// ParameterSetProvider interface for managing sets of parameters.
type ParameterSetProvider interface {
	GetDataStore() Store

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
