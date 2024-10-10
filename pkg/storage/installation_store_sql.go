package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"

	"get.porter.sh/porter/pkg/tracing"
)

var _ InstallationProvider = &InstallationStoreSQL{}

// InstallationStoreSQL is a persistent store for installation documents.
type InstallationStoreSQL struct {
	db      *gorm.DB
	encrypt EncryptionHandler
	decrypt EncryptionHandler
}

// NewInstallationStoreSQL creates a new InstallationStoreSQL
func NewInstallationStoreSQL(db *gorm.DB) *InstallationStoreSQL {
	return &InstallationStoreSQL{
		db:      db,
		encrypt: noOpEncryptionHandler,
		decrypt: noOpEncryptionHandler,
	}
}

// EnsureInstallationIndicesSQL creates indices on the installations table.
// TODO move to gorm migration step
func EnsureInstallationIndicesSQL(ctx context.Context, db *gorm.DB) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debug("Initializing installation table indices")

	// Create indices similar to MongoDB
	err := db.WithContext(ctx).Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_installations_namespace_name ON installations (namespace, name);
		CREATE INDEX IF NOT EXISTS idx_runs_namespace_installation ON runs (namespace, installation);
		CREATE INDEX IF NOT EXISTS idx_results_namespace_installation ON results (namespace, installation);
		CREATE INDEX IF NOT EXISTS idx_results_run_id ON results (run_id);
		CREATE INDEX IF NOT EXISTS idx_outputs_namespace_installation_result_id ON outputs (namespace, installation, result_id DESC);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_outputs_result_id_name ON outputs (result_id, name);
		CREATE INDEX IF NOT EXISTS idx_outputs_namespace_installation_name_result_id ON outputs (namespace, installation, name, result_id DESC);
	`).Error

	return span.Error(err)
}

func (s *InstallationStoreSQL) ListInstallations(ctx context.Context, listOptions ListOptions) ([]Installation, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var installations []Installation
	query := s.db.WithContext(ctx).Order("namespace, name")

	// Filter by Namespace
	if listOptions.Namespace != "" {
		query = query.Where("namespace = ?", listOptions.Namespace)
	}

	// Filter by Name
	if listOptions.Name != "" {
		query = query.Where("name = ?", listOptions.Name)
	}

	// Filter by Labels
	if len(listOptions.Labels) > 0 {
		for key, value := range listOptions.Labels {
			query = query.Where("labels->? = ?", key, value)
		}
	}

	// Apply Skip (Offset)
	if listOptions.Skip > 0 {
		query = query.Offset(int(listOptions.Skip))
	}

	// Apply Limit
	if listOptions.Limit > 0 {
		query = query.Limit(int(listOptions.Limit))
	}

	// Execute the query
	err := query.Find(&installations).Error
	if err != nil {
		return nil, log.Error(err)
	}

	return installations, nil
}

func (s *InstallationStoreSQL) ListRuns(ctx context.Context, namespace string, installation string) ([]Run, map[string][]Result, error) {
	var runs []Run
	var results []Result

	err := s.db.WithContext(ctx).Where("namespace = ? AND installation = ?", namespace, installation).Order("id").Find(&runs).Error
	if err != nil {
		return nil, nil, err
	}

	err = s.db.WithContext(ctx).Where("namespace = ? AND installation = ?", namespace, installation).Find(&results).Error
	if err != nil {
		return runs, nil, err
	}

	resultsMap := make(map[string][]Result, len(runs))
	for _, run := range runs {
		resultsMap[run.ID] = []Result{}
	}
	for _, res := range results {
		if _, ok := resultsMap[res.RunID]; ok {
			resultsMap[res.RunID] = append(resultsMap[res.RunID], res)
		}
	}

	return runs, resultsMap, nil
}

func (s *InstallationStoreSQL) ListResults(ctx context.Context, runID string) ([]Result, error) {
	var results []Result
	err := s.db.WithContext(ctx).Where("run_id = ?", runID).Order("id").Find(&results).Error
	return results, err
}

func (s *InstallationStoreSQL) ListOutputs(ctx context.Context, resultID string) ([]Output, error) {
	var outputs []Output
	err := s.db.WithContext(ctx).Where("result_id = ?", resultID).Order("result_id, name").Find(&outputs).Error
	return outputs, err
}

func (s *InstallationStoreSQL) FindInstallations(ctx context.Context, findOpts FindOptions) ([]Installation, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var installations []Installation
	query := s.db.WithContext(ctx)

	// Apply filters
	for key, value := range findOpts.Filter {
		switch key {
		case "namespace":
			query = query.Where("namespace = ?", value)
		case "name":
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		default:
			// Assume it's a label query
			if strings.HasPrefix(key, "labels.") {
				labelKey := strings.TrimPrefix(key, "labels.")
				query = query.Where("labels->? @> ?", labelKey, fmt.Sprintf("\"%v\"", value))
			}
		}
	}

	// Apply sorting
	for _, sortField := range findOpts.Sort {
		if strings.HasPrefix(sortField, "-") {
			query = query.Order(strings.TrimPrefix(sortField, "-") + " DESC")
		} else {
			query = query.Order(sortField)
		}
	}

	// Apply projection
	if len(findOpts.Select) > 0 {
		var selectFields []string
		for _, field := range findOpts.Select {
			if field.Value.(bool) {
				selectFields = append(selectFields, field.Key)
			}
		}
		if len(selectFields) > 0 {
			query = query.Select(selectFields)
		} else {
			// If all fields are set to false, we need to select all fields except those
			query = query.Omit(getOmitFields(findOpts.Select)...)
		}
	}

	// Apply Skip (Offset)
	if findOpts.Skip > 0 {
		query = query.Offset(int(findOpts.Skip))
	}

	// Apply Limit
	if findOpts.Limit > 0 {
		query = query.Limit(int(findOpts.Limit))
	}

	// Execute the query
	err := query.Find(&installations).Error
	if err != nil {
		return nil, log.Error(err)
	}

	return installations, nil
}

// Helper function to get fields to omit
func getOmitFields(select_ bson.D) []string {
	var omitFields []string
	for _, field := range select_ {
		if !field.Value.(bool) {
			omitFields = append(omitFields, field.Key)
		}
	}
	return omitFields
}

func (s *InstallationStoreSQL) GetInstallation(ctx context.Context, namespace string, name string) (Installation, error) {
	var installation Installation
	err := s.db.WithContext(ctx).Where("namespace = ? AND name = ?", namespace, name).First(&installation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Installation{}, ErrNotFound{Collection: "installations"}
	}
	return installation, err
}

func (s *InstallationStoreSQL) GetRun(ctx context.Context, id string) (Run, error) {
	var run Run
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Run{}, ErrNotFound{Collection: "runs"}
	}
	return run, err
}

func (s *InstallationStoreSQL) GetResult(ctx context.Context, id string) (Result, error) {
	var result Result
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Result{}, ErrNotFound{Collection: "results"}
	}
	return result, err
}

func (s *InstallationStoreSQL) GetLastRun(ctx context.Context, namespace string, installation string) (Run, error) {
	var run Run
	err := s.db.WithContext(ctx).Where("namespace = ? AND installation = ?", namespace, installation).Order("id DESC").First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Run{}, ErrNotFound{Collection: "runs"}
	}
	return run, err
}

func (s *InstallationStoreSQL) GetLastOutput(ctx context.Context, namespace string, installation string, name string) (Output, error) {
	var output Output
	err := s.db.WithContext(ctx).
		Where("namespace = ? AND installation = ? AND name = ?", namespace, installation, name).
		Order("result_id DESC").
		First(&output).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Output{}, ErrNotFound{Collection: "outputs", Item: name}
	}
	return output, err
}

func (s *InstallationStoreSQL) GetLastOutputs(ctx context.Context, namespace string, installation string) (Outputs, error) {
	var outputs []Output

	db := s.db.WithContext(ctx)

	subQuery := db.Table("outputs").
		Select("name, MAX(result_id) as max_result_id").
		Where("namespace = ? AND installation = ?", namespace, installation).
		Group("name")

	err := db.Table("outputs").
		Select("outputs.*").
		Joins("JOIN (?) AS latest ON outputs.name = latest.name AND outputs.result_id = latest.max_result_id", subQuery).
		Where("outputs.namespace = ? AND outputs.installation = ?", namespace, installation).
		Find(&outputs).Error

	if err != nil {
		return Outputs{}, err
	}

	return NewOutputs(outputs), err
}

func (s *InstallationStoreSQL) GetLogs(ctx context.Context, runID string) (string, bool, error) {
	var output Output
	err := s.db.WithContext(ctx).
		Where("run_id = ? AND name = ?", runID, "io.cnab.outputs.invocationImageLogs").
		Order("result_id").
		First(&output).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	return string(output.Value), err == nil, err
}

func (s *InstallationStoreSQL) GetLastLogs(ctx context.Context, namespace string, installation string) (string, bool, error) {
	var output Output
	err := s.db.WithContext(ctx).
		Where("namespace = ? AND installation = ? AND name = ?", namespace, installation, "io.cnab.outputs.invocationImageLogs").
		Order("result_id DESC").
		First(&output).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	return string(output.Value), err == nil, err
}

func (s *InstallationStoreSQL) InsertInstallation(ctx context.Context, installation Installation) error {
	installation.SchemaVersion = DefaultInstallationSchemaVersion
	return s.db.WithContext(ctx).Create(&installation).Error
}

func (s *InstallationStoreSQL) InsertRun(ctx context.Context, run Run) error {
	return s.db.WithContext(ctx).Create(&run).Error
}

func (s *InstallationStoreSQL) InsertResult(ctx context.Context, result Result) error {
	return s.db.WithContext(ctx).Create(&result).Error
}

func (s *InstallationStoreSQL) InsertOutput(ctx context.Context, output Output) error {
	return s.db.WithContext(ctx).Create(&output).Error
}

func (s *InstallationStoreSQL) UpdateInstallation(ctx context.Context, installation Installation) error {
	installation.SchemaVersion = DefaultInstallationSchemaVersion
	return s.db.WithContext(ctx).Save(&installation).Error
}

func (s *InstallationStoreSQL) UpsertRun(ctx context.Context, run Run) error {
	return s.db.WithContext(ctx).Save(&run).Error
}

func (s *InstallationStoreSQL) UpsertInstallation(ctx context.Context, installation Installation) error {
	installation.SchemaVersion = DefaultInstallationSchemaVersion
	return s.db.WithContext(ctx).Save(&installation).Error
}

func (s *InstallationStoreSQL) RemoveInstallation(ctx context.Context, namespace string, name string) error {
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("namespace = ? AND name = ?", namespace, name).Delete(&Installation{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("namespace = ? AND installation = ?", namespace, name).Delete(&Run{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("namespace = ? AND installation = ?", namespace, name).Delete(&Result{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("namespace = ? AND installation = ?", namespace, name).Delete(&Output{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
