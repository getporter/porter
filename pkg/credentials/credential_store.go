package credentials

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

var _ Provider = &CredentialStore{}

const (
	CollectionCredentials = "credentials"
)

// CredentialStore provides access to credential sets by instantiating plugins that
// implement CRUD storage.
type CredentialStore struct {
	Documents storage.Store
	Secrets   secrets.Store
}

func NewCredentialStore(storage storage.Store, secrets secrets.Store) *CredentialStore {
	return &CredentialStore{
		Documents: storage,
		Secrets:   secrets,
	}
}

// Initialize the underlying storage with any additional schema changes, such as indexes.
func (s CredentialStore) Initialize() error {
	indices := storage.EnsureIndexOptions{
		Indices: []storage.Index{
			// query credentials by namespace + name
			{Collection: CollectionCredentials, Keys: []string{"namespace", "name"}, Unique: true},
		},
	}
	return s.Documents.EnsureIndex(indices)
}

func (s CredentialStore) GetDataStore() storage.Store {
	return s.Documents
}

/*
	Secrets
*/

func (s CredentialStore) ResolveAll(creds CredentialSet) (secrets.Set, error) {
	resolvedCreds := make(secrets.Set)
	var resolveErrors error

	for _, cred := range creds.Credentials {
		value, err := s.Secrets.Resolve(cred.Source.Key, cred.Source.Value)
		if err != nil {
			resolveErrors = multierror.Append(resolveErrors, errors.Wrapf(err, "unable to resolve credential %s.%s from %s %s", creds.Name, cred.Name, cred.Source.Key, cred.Source.Value))
		}

		resolvedCreds[cred.Name] = value
	}

	return resolvedCreds, resolveErrors
}

func (s CredentialStore) Validate(creds CredentialSet) error {
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

func (s CredentialStore) InsertCredentialSet(cset CredentialSet) error {
	opts := storage.InsertOptions{
		Documents: []interface{}{cset},
	}
	return s.Documents.Insert(CollectionCredentials, opts)
}

func (s CredentialStore) ListCredentialSets(namespace string, name string, labels map[string]string) ([]CredentialSet, error) {
	var out []CredentialSet
	opts := storage.FindOptions{
		Filter: storage.CreateListFiler(namespace, name, labels),
	}
	err := s.Documents.Find(CollectionCredentials, opts, &out)
	return out, err
}

func (s CredentialStore) GetCredentialSet(namespace string, name string) (CredentialSet, error) {
	var out CredentialSet
	opts := storage.FindOptions{
		Filter: map[string]interface{}{
			"namespace": namespace,
			"name":      name,
		},
	}
	err := s.Documents.FindOne(CollectionCredentials, opts, &out)
	return out, err
}

func (s CredentialStore) UpdateCredentialSet(creds CredentialSet) error {
	opts := storage.UpdateOptions{
		Document: creds,
	}
	return s.Documents.Update(CollectionCredentials, opts)
}

func (s CredentialStore) UpsertCredentialSet(creds CredentialSet) error {
	opts := storage.UpdateOptions{
		Document: creds,
		Upsert:   true,
	}
	return s.Documents.Update(CollectionCredentials, opts)
}

func (s CredentialStore) RemoveCredentialSet(namespace string, name string) error {
	opts := storage.RemoveOptions{
		Namespace: namespace,
		Name:      name,
	}
	return s.Documents.Remove(CollectionCredentials, opts)
}
