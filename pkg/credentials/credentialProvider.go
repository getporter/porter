package credentials

import (
	"github.com/cnabio/cnab-go/credentials"
)

// CredentialProvider interface for managing sets of credentials.
type CredentialProvider interface {
	CredentialStore
	ResolveAll(creds credentials.CredentialSet) (credentials.Set, error)
	Validate(credentials.CredentialSet) error
}

// CredentialStore is an interface representing cnab-go's credentials.Store
type CredentialStore interface {
	List() ([]string, error)
	Save(credentials.CredentialSet) error
	Read(name string) (credentials.CredentialSet, error)
	ReadAll() ([]credentials.CredentialSet, error)
	Delete(name string) error
}
