package storage

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

var _ CredentialSetProvider = &CredentialStore{}

const (
	CollectionCredentials = "credentials"
)

// CredentialStore provides access to work with Porter
// credential sets, usually referred to as "credentials" as a shorthand.
//
// Credential Sets define mappings from a credential needed by a bundle to where
// to look for it when the bundle is run. For example: Bundle needs Azure
// storage connection string and it should look for it in an environment
// variable named `AZURE_STORATE_CONNECTION_STRING` or a key named `dev-conn`.
//
// Porter discourages storing the value of the credential directly, though it
// it is possible. Instead Porter encourages the best practice of defining
// mappings in the credential sets, and then storing the values in secret stores
// such as a key/value store like Hashicorp Vault, or Azure Key Vault.
// See the get.porter.sh/porter/pkg/secrets package for more on how Porter
// handles accessing secrets.
type CredentialStore struct {
	Documents Store
	Secrets   secrets.Store
}

func NewCredentialStore(storage Store, secrets secrets.Store) *CredentialStore {
	return &CredentialStore{
		Documents: storage,
		Secrets:   secrets,
	}
}

// Initialize the underlying storage with any additional schema changes, such as indexes.
func (s CredentialStore) Initialize(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debug("Initializing credentials collection indices")

	indices := EnsureIndexOptions{
		Indices: []Index{
			// query credentials by namespace + name
			{Collection: CollectionCredentials, Keys: []string{"namespace", "name"}, Unique: true},
		},
	}
	err := s.Documents.EnsureIndex(ctx, indices)
	return span.Error(err)
}

func (s CredentialStore) GetDataStore() Store {
	return s.Documents
}

/*
	Secrets
*/

func (s CredentialStore) ResolveAll(ctx context.Context, creds CredentialSet) (secrets.Set, error) {
	resolvedCreds := make(secrets.Set)
	var resolveErrors error

	for _, cred := range creds.Credentials {
		value, err := s.Secrets.Resolve(ctx, cred.Source.Key, cred.Source.Value)
		if err != nil {
			resolveErrors = multierror.Append(resolveErrors, errors.Wrapf(err, "unable to resolve credential %s.%s from %s %s", creds.Name, cred.Name, cred.Source.Key, cred.Source.Value))
		}

		resolvedCreds[cred.Name] = value
	}

	return resolvedCreds, resolveErrors
}

func (s CredentialStore) Validate(ctx context.Context, creds CredentialSet) error {
	validSources := []string{secrets.SourceSecret, host.SourceValue, host.SourceEnv, host.SourcePath, host.SourceCommand}
	var errors error

	for _, cs := range creds.Credentials {
		valid := false
		for _, validSource := range validSources {
			if cs.Source.Key == validSource {
				valid = true
				break
			}
		}
		if valid == false {
			errors = multierror.Append(errors, fmt.Errorf(
				"%s is not a valid source. Valid sources are: %s",
				cs.Source.Key,
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
	creds.SchemaVersion = CredentialSetSchemaVersion
	opts := InsertOptions{
		Documents: []interface{}{creds},
	}
	return s.Documents.Insert(ctx, CollectionCredentials, opts)
}

func (s CredentialStore) ListCredentialSets(ctx context.Context, namespace string, name string, labels map[string]string) ([]CredentialSet, error) {
	var out []CredentialSet
	opts := FindOptions{
		Filter: CreateListFiler(namespace, name, labels),
	}
	err := s.Documents.Find(ctx, CollectionCredentials, opts, &out)
	return out, err
}

func (s CredentialStore) GetCredentialSet(ctx context.Context, namespace string, name string) (CredentialSet, error) {
	var out CredentialSet
	opts := FindOptions{
		Filter: map[string]interface{}{
			"namespace": namespace,
			"name":      name,
		},
	}
	err := s.Documents.FindOne(ctx, CollectionCredentials, opts, &out)
	return out, err
}

func (s CredentialStore) UpdateCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = CredentialSetSchemaVersion
	opts := UpdateOptions{
		Document: creds,
	}
	return s.Documents.Update(ctx, CollectionCredentials, opts)
}

func (s CredentialStore) UpsertCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = CredentialSetSchemaVersion
	opts := UpdateOptions{
		Document: creds,
		Upsert:   true,
	}
	return s.Documents.Update(ctx, CollectionCredentials, opts)
}

func (s CredentialStore) RemoveCredentialSet(ctx context.Context, namespace string, name string) error {
	opts := RemoveOptions{
		Namespace: namespace,
		Name:      name,
	}
	return s.Documents.Remove(ctx, CollectionCredentials, opts)
}
