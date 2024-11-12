package storage

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"get.porter.sh/porter/pkg/secrets"
	hostSecrets "get.porter.sh/porter/pkg/secrets/plugins/host"
	"get.porter.sh/porter/pkg/tracing"
)

var _ CredentialSetProvider = (*CredentialStoreSQL)(nil)

// CredentialStoreSQL is a wrapper around Porter's datastore
// providing typed access and additional business logic around
// credential sets, usually referred to as "credentials" as a shorthand.
type CredentialStoreSQL struct {
	db          *gorm.DB
	Secrets     secrets.Store
	HostSecrets hostSecrets.Store
}

func NewCredentialStoreSQL(db *gorm.DB, secrets secrets.Store) *CredentialStoreSQL {
	return &CredentialStoreSQL{
		db:          db,
		Secrets:     secrets,
		HostSecrets: hostSecrets.NewStore(),
	}
}

/*
	Secrets
*/

func (s CredentialStoreSQL) ResolveAll(ctx context.Context, creds CredentialSet) (secrets.Set, error) {
	return resolveAll(ctx, creds.Credentials, s.HostSecrets, s.Secrets, creds.Name, "credential")
}

func (s CredentialStoreSQL) Validate(ctx context.Context, creds CredentialSet) error {
	return validate(creds.Credentials)
}

/*
  Document Storage
*/

func (s CredentialStoreSQL) InsertCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = DefaultCredentialSetSchemaVersion
	return s.db.WithContext(ctx).Create(&creds).Error
}

func (s CredentialStoreSQL) ListCredentialSets(ctx context.Context, listOptions ListOptions) ([]CredentialSet, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var out []CredentialSet

	// If no filters are provided, return an empty list
	if listOptions.Namespace == "" && listOptions.Name == "" && len(listOptions.Labels) == 0 {
		return out, nil
	}

	query := s.db.WithContext(ctx).Order("namespace ASC, name ASC")

	if listOptions.Namespace != "" && listOptions.Namespace != "*" {
		query = query.Where("namespace = ?", listOptions.Namespace)
	}

	if listOptions.Name != "" {
		query = query.Where("name LIKE ?", "%"+listOptions.Name+"%")
	}

	// Filter by Labels
	if len(listOptions.Labels) > 0 {
		for key, value := range listOptions.Labels {
			query = query.Where("labels->? = ?", key, value)
		}
	}

	if listOptions.Skip > 0 {
		query = query.Offset(int(listOptions.Skip))
	}

	if listOptions.Limit > 0 {
		query = query.Limit(int(listOptions.Limit))
	}

	err := query.Find(&out).Error
	if err != nil {
		return nil, log.Error(err)
	}
	return out, err
}

func (s CredentialStoreSQL) GetCredentialSet(ctx context.Context, namespace string, name string) (CredentialSet, error) {
	var out CredentialSet
	err := s.db.WithContext(ctx).
		Where("namespace = ? AND name = ?", namespace, name).
		First(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CredentialSet{}, ErrNotFound{Collection: "credentials"}
	}
	return out, err
}
func (s CredentialStoreSQL) FindCredentialSet(ctx context.Context, namespace string, name string) (CredentialSet, error) {
	var out CredentialSet

	query := s.db.WithContext(ctx).
		Where("name = ?", name).
		Where("(namespace = ? OR namespace = '' OR namespace IS NULL)", namespace).
		Order("namespace DESC") // DESC ensures that the namespace is prioritized over the empty or null

	err := query.First(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CredentialSet{}, ErrNotFound{Collection: "credentials"}
	}
	return out, err
}

func (s CredentialStoreSQL) UpdateCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = DefaultCredentialSetSchemaVersion
	return s.db.WithContext(ctx).
		Where("namespace = ? AND name = ?", creds.Namespace, creds.Name).
		Save(&creds).Error
}

func (s CredentialStoreSQL) UpsertCredentialSet(ctx context.Context, creds CredentialSet) error {
	creds.SchemaVersion = DefaultCredentialSetSchemaVersion
	return s.db.WithContext(ctx).Save(&creds).Error
}
func (s CredentialStoreSQL) RemoveCredentialSet(ctx context.Context, namespace string, name string) error {
	return s.db.WithContext(ctx).Where("namespace = ? AND name = ?", namespace, name).Delete(&CredentialSet{}).Error
}
