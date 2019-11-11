package credentials

import (
	"github.com/cnabio/cnab-go/credentials"
)

// CredentialProvider interface for managing sets of credentials.
type CredentialProvider interface {
	ICredentialStore
	ResolveAll(creds credentials.CredentialSet) (credentials.Set, error)
}

type ICredentialStore interface {
	List() ([]string, error)
	Save(credentials.CredentialSet) error
	Read(name string) (credentials.CredentialSet, error)
	ReadAll() ([]credentials.CredentialSet, error)
	Delete(name string) error
}