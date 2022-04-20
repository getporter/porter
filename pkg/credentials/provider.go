package credentials

import (
	"context"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
)

// Provider is Porter's interface for managing and resolving credentials.
type Provider interface {
	GetDataStore() storage.Store
	ResolveAll(ctx context.Context, creds CredentialSet) (secrets.Set, error)
	Validate(ctx context.Context, creds CredentialSet) error
	InsertCredentialSet(ctx context.Context, creds CredentialSet) error
	ListCredentialSets(ctx context.Context, namespace string, name string, labels map[string]string) ([]CredentialSet, error)
	GetCredentialSet(ctx context.Context, namespace string, name string) (CredentialSet, error)
	UpdateCredentialSet(ctx context.Context, creds CredentialSet) error
	RemoveCredentialSet(ctx context.Context, namespace string, name string) error
	UpsertCredentialSet(ctx context.Context, creds CredentialSet) error
}
