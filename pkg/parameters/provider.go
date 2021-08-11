package parameters

import (
	"get.porter.sh/porter/pkg/secrets"
)

// Provider interface for managing sets of parameters.
type Provider interface {
	// ResolveAll parameter values in the parameter set.
	ResolveAll(params ParameterSet) (secrets.Set, error)

	// Validate the parameter set is defined properly.
	Validate(params ParameterSet) error

	InsertParameterSet(params ParameterSet) error
	ListParameterSets(namespace string, name string, labels map[string]string) ([]ParameterSet, error)
	GetParameterSet(namespace string, name string) (ParameterSet, error)
	UpdateParameterSet(params ParameterSet) error
	UpsertParameterSet(params ParameterSet) error
	RemoveParameterSet(namespace string, name string) error
}
