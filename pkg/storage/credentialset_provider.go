package storage

import (
	"context"

	"get.porter.sh/porter/pkg/secrets"
)

// CredentialSetProvider is Porter's interface for managing and resolving credentials.
type CredentialSetProvider interface {
	// FindCredentialSet finds a credential set by name in the specified namespace, or falls back to the global namespace.
	FindCredentialSet(ctx context.Context, namespace string, name string) (CredentialSet, error)

	// ResolveAll credential values in the credential set.
	ResolveAll(ctx context.Context, creds CredentialSet) (secrets.Set, error)

	// Validate the credential set is defined properly.
	Validate(ctx context.Context, creds CredentialSet) error

	InsertCredentialSet(ctx context.Context, creds CredentialSet) error
	ListCredentialSets(ctx context.Context, listOptions ListOptions) ([]CredentialSet, error)
	GetCredentialSet(ctx context.Context, namespace string, name string) (CredentialSet, error)
	UpdateCredentialSet(ctx context.Context, creds CredentialSet) error
	UpsertCredentialSet(ctx context.Context, creds CredentialSet) error
	RemoveCredentialSet(ctx context.Context, namespace string, name string) error
}
