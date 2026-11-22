package storage

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	hostSecrets "get.porter.sh/porter/pkg/secrets/plugins/host"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/hashicorp/go-multierror"
)

var _ CredentialSetProvider = &CredentialStore{}

const (
	CollectionCredentials = "credentials"
)

// CredentialStore is a wrapper around Porter's datastore
// providing typed access and additional business logic around
// credential sets, usually referred to as "credentials" as a shorthand.
type CredentialStore struct {
	Documents   BackendClient
	Secrets     secrets.Store
	HostSecrets hostSecrets.Store
}

func NewCredentialStore(storage BackendClient, secrets secrets.Store) *CredentialStore {
	return &CredentialStore{
		Documents:   storage,
		Secrets:     secrets,
		HostSecrets: hostSecrets.NewStore(),
	}
}

/*
	Secrets
*/

func (s CredentialStore) ResolveAll(ctx context.Context, creds CredentialSet, keys []string) (secrets.Set, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	resolvedCreds := make(secrets.Set)
	var resolveErrors *multierror.Error

	for _, key := range keys {
		cred, ok := creds.GetCredential(key)
		if !ok {
			resolveErrors = multierror.Append(resolveErrors, span.Errorf("credential %s not found", key))
			continue
		}

		var value string
		var err error
		if isHandledByHostPlugin(cred.Source.Strategy) {
			value, err = s.HostSecrets.Resolve(ctx, cred.Source.Strategy, cred.Source.Hint)
		} else {
			value, err = s.Secrets.Resolve(ctx, cred.Source.Strategy, cred.Source.Hint)
		}
		if err != nil {
			resolveErrors = multierror.Append(resolveErrors, span.Errorf("unable to resolve credential %s.%s from %s %s: %w", creds.Name, cred.Name, cred.Source.Strategy, cred.Source.Hint, err))
		}

		resolvedCreds[cred.Name] = value
	}

	return resolvedCreds, resolveErrors.ErrorOrNil()
}

func (s CredentialStore) Validate(ctx context.Context, creds CredentialSet) error {
	validSources := []string{secrets.SourceSecret, host.SourceValue, host.SourceEnv, host.SourcePath, host.SourceCommand}
	var errors error

	for _, cs := range creds.Credentials {
		valid := false
		for _, validSource := range validSources {
			if cs.Source.Strategy == validSource {
				valid = true
				break
			}
		}
		if !valid {
			errors = multierror.Append(errors, fmt.Errorf(
				"%s is not a valid source. Valid sources are: %s",
				cs.Source.Strategy,
				strings.Join(validSources, ", "),
			))
		}
	}

	return errors
}

/*
  Document Storage
*/

func (s CredentialStore) InsertCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = DefaultCredentialSetSchemaVersion
	return s.Documents.put(ctx, protov4.PutRequest_INSERT, creds)
}

func (s CredentialStore) ListCredentialSets(ctx context.Context, listOptions ListOptions) ([]CredentialSet, error) {
	return listDocuments[CredentialSet](ctx, s.Documents, listOptions.toListRequest(protov4.Kind_KIND_CREDENTIAL_SET))
}

func (s CredentialStore) GetCredentialSet(ctx context.Context, namespace string, name string) (CredentialSet, error) {
	return getDocument[CredentialSet](ctx, s.Documents, &protov4.GetRequest{
		Kind:      protov4.Kind_KIND_CREDENTIAL_SET,
		Namespace: namespace,
		Name:      name,
	})
}

// GetCredentialSetWithFallback retrieves the credential set from the specified
// namespace, falling back to the global namespace when it isn't found there.
func (s CredentialStore) GetCredentialSetWithFallback(ctx context.Context, namespace string, name string) (CredentialSet, error) {
	return getDocument[CredentialSet](ctx, s.Documents, &protov4.GetRequest{
		Kind:             protov4.Kind_KIND_CREDENTIAL_SET,
		Namespace:        namespace,
		Name:             name,
		FallbackToGlobal: true,
	})
}

func (s CredentialStore) UpdateCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = DefaultCredentialSetSchemaVersion
	return s.Documents.put(ctx, protov4.PutRequest_UPDATE, creds)
}

func (s CredentialStore) UpsertCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = DefaultCredentialSetSchemaVersion
	return s.Documents.put(ctx, protov4.PutRequest_UPSERT, creds)
}

func (s CredentialStore) RemoveCredentialSet(ctx context.Context, namespace string, name string) error {
	_, err := s.Documents.backend.Remove(ctx, &protov4.RemoveRequest{
		Kind:      protov4.Kind_KIND_CREDENTIAL_SET,
		Namespace: namespace,
		Name:      name,
	})
	return err
}
