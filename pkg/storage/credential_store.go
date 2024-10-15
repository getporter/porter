package storage

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"get.porter.sh/porter/pkg/secrets"
	hostSecrets "get.porter.sh/porter/pkg/secrets/plugins/host"
	"get.porter.sh/porter/pkg/tracing"
)

var _ CredentialSetProvider = &CredentialStore{}

const (
	CollectionCredentials = "credentials"
)

// CredentialStore is a wrapper around Porter's datastore
// providing typed access and additional business logic around
// credential sets, usually referred to as "credentials" as a shorthand.
type CredentialStore struct {
	Documents   Store
	Secrets     secrets.Store
	HostSecrets hostSecrets.Store
}

func NewCredentialStore(storage Store, secrets secrets.Store) *CredentialStore {
	return &CredentialStore{
		Documents:   storage,
		Secrets:     secrets,
		HostSecrets: hostSecrets.NewStore(),
	}
}

// EnsureCredentialIndices creates indices on the credentials collection.
func EnsureCredentialIndices(ctx context.Context, store Store) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debug("Initializing credentials collection indices")

	indices := EnsureIndexOptions{
		Indices: []Index{
			// query credentials by namespace + name
			{Collection: CollectionCredentials, Keys: []string{"namespace", "name"}, Unique: true},
		},
	}
	err := store.EnsureIndex(ctx, indices)
	return span.Error(err)
}

func (s CredentialStore) FindCredentialSet(ctx context.Context, namespace string, name string) (CredentialSet, error) {
	var out CredentialSet
	// Try to get the creds in the local namespace first, fallback to the global creds
	query := FindOptions{
		Sort: []string{"-namespace"},
		Filter: bson.M{
			"name": name,
			"$or": []bson.M{
				{"namespace": ""},
				{"namespace": namespace},
			},
		},
	}
	err := s.Documents.FindOne(ctx, CollectionCredentials, query, &out)
	return out, err
}

/*
	Secrets
*/

func (s CredentialStore) ResolveAll(ctx context.Context, creds CredentialSet) (secrets.Set, error) {
	return resolveAll(ctx, creds.Credentials, s.HostSecrets, s.Secrets, creds.Name, "credential")
}

func (s CredentialStore) Validate(ctx context.Context, creds CredentialSet) error {
	return validate(creds.Credentials)
}

/*
  Document Storage
*/

func (s CredentialStore) InsertCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = DefaultCredentialSetSchemaVersion
	opts := InsertOptions{
		Documents: []interface{}{creds},
	}
	return s.Documents.Insert(ctx, CollectionCredentials, opts)
}

func (s CredentialStore) ListCredentialSets(ctx context.Context, listOptions ListOptions) ([]CredentialSet, error) {
	var out []CredentialSet
	err := s.Documents.Find(ctx, CollectionCredentials, listOptions.ToFindOptions(), &out)
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
	creds.SchemaVersion = DefaultCredentialSetSchemaVersion
	opts := UpdateOptions{
		Document: creds,
	}
	return s.Documents.Update(ctx, CollectionCredentials, opts)
}

func (s CredentialStore) UpsertCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = DefaultCredentialSetSchemaVersion
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
