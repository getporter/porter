package parameters

import "github.com/cnabio/cnab-go/valuesource"

// TODO: clone of credentialProvider from credentials pkg
// Generalize to DRY out?

// ParameterProvider interface for managing sets of parameters.
type ParameterProvider interface {
	ParameterStore
	ResolveAll(creds ParameterSet) (valuesource.Set, error)
	Validate(ParameterSet) error
}

// ParameterStore is an interface representing parameters.Store
type ParameterStore interface {
	List() ([]string, error)
	Save(ParameterSet) error
	Read(name string) (ParameterSet, error)
	ReadAll() ([]ParameterSet, error)
	Delete(name string) error
}
