package storage

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"get.porter.sh/porter/pkg/secrets"
	hostSecrets "get.porter.sh/porter/pkg/secrets/plugins/host"
	"get.porter.sh/porter/pkg/tracing"
)

var _ ParameterSetProvider = (*ParameterStoreSQL)(nil)

// ParameterStoreSQL provides access to parameter sets by instantiating plugins that
// implement CRUD storage.
type ParameterStoreSQL struct {
	db          *gorm.DB
	Secrets     secrets.Store
	HostSecrets hostSecrets.Store
}

func NewParameterStoreSQL(db *gorm.DB, secrets secrets.Store) *ParameterStoreSQL {
	return &ParameterStoreSQL{
		db:          db,
		Secrets:     secrets,
		HostSecrets: hostSecrets.NewStore(),
	}
}

func (s ParameterStoreSQL) ResolveAll(ctx context.Context, params ParameterSet) (secrets.Set, error) {
	return resolveAll(ctx, params.Parameters, s.HostSecrets, s.Secrets, params.Name, "parameter")
}

func (s ParameterStoreSQL) Validate(ctx context.Context, params ParameterSet) error {
	return validate(params.Parameters)
}
func (s ParameterStoreSQL) InsertParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = DefaultParameterSetSchemaVersion
	return s.db.WithContext(ctx).Create(&params).Error
}
func (s ParameterStoreSQL) ListParameterSets(ctx context.Context, listOptions ListOptions) ([]ParameterSet, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var out []ParameterSet

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
func (s ParameterStoreSQL) GetParameterSet(ctx context.Context, namespace string, name string) (ParameterSet, error) {
	var out ParameterSet
	err := s.db.WithContext(ctx).Where("namespace = ? AND name = ?", namespace, name).First(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ParameterSet{}, ErrNotFound{Collection: "parameters"}
	}
	return out, err
}

func (s ParameterStoreSQL) FindParameterSet(ctx context.Context, namespace string, name string) (ParameterSet, error) {
	var out ParameterSet

	query := s.db.WithContext(ctx).
		Where("name = ?", name).
		Where("(namespace = ? OR namespace = '' OR namespace IS NULL)", namespace).
		Order("namespace DESC") // DESC ensures that the namespace is prioritized over the empty or null

	err := query.First(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ParameterSet{}, ErrNotFound{Collection: "parameters"}
	}
	return out, err
}

func (s ParameterStoreSQL) UpdateParameterSet(ctx context.Context, params ParameterSet) error {
	return s.db.WithContext(ctx).
		Where("namespace = ? AND name = ?", params.Namespace, params.Name).
		Save(&params).Error
}
func (s ParameterStoreSQL) UpsertParameterSet(ctx context.Context, params ParameterSet) error {
	params.SchemaVersion = DefaultParameterSetSchemaVersion
	return s.db.WithContext(ctx).Save(&params).Error
}
func (s ParameterStoreSQL) RemoveParameterSet(ctx context.Context, namespace string, name string) error {
	return s.db.WithContext(ctx).Where("namespace = ? AND name = ?", namespace, name).Delete(&ParameterSet{}).Error
}
