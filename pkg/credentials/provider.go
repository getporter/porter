package credentials

import (
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
)

// Provider is Porter's interface for managing and resolving credentials.
type Provider interface {
	GetDataStore() storage.Store
	ResolveAll(creds CredentialSet) (secrets.Set, error)
	Validate(creds CredentialSet) error
	InsertCredentialSet(creds CredentialSet) error
	ListCredentialSets(namespace string, name string, labels map[string]string) ([]CredentialSet, error)
	GetCredentialSet(namespace string, name string) (CredentialSet, error)
	UpdateCredentialSet(creds CredentialSet) error
	RemoveCredentialSet(namespace string, name string) error
	UpsertCredentialSet(creds CredentialSet) error
}
