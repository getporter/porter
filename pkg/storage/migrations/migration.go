package migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/migrations/crudstore"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel/attribute"
)

// Migration can connect to a legacy Porter v0.38 storage plugin migrate the data
// in the specified account into a target account compatible with the current
// version of Porter.
type Migration struct {
	config      *config.Config
	opts        storage.MigrateOptions
	sourceStore crudstore.Store
	destStore   storage.Store
	sanitizer   *storage.Sanitizer
	pluginConn  *pluggable.PluginConnection
}

func NewMigration(c *config.Config, opts storage.MigrateOptions, destStore storage.Store, sanitizer *storage.Sanitizer) *Migration {
	return &Migration{
		config:    c,
		opts:      opts,
		destStore: destStore,
		sanitizer: sanitizer,
	}
}

// Connect loads the legacy plugin specified by the source storage account.
func (m *Migration) Connect(ctx context.Context) error {
	ctx, log := tracing.StartSpan(ctx,
		attribute.String("storage-name", m.opts.SourceAccount))
	defer log.EndSpan()

	// Create a config file that uses the old PORTER_HOME
	oldConfig := config.New()
	oldConfig.SetHomeDir(m.opts.OldHome)
	oldConfig.SetPorterPath(filepath.Join(m.opts.OldHome, "porter"))
	oldConfig.Load(ctx, nil)
	oldConfig.Setenv(config.EnvHOME, m.opts.OldHome)

	l := pluggable.NewPluginLoader(oldConfig)
	conn, err := l.Load(ctx, m.LegacyStoragePluginConfig())
	if err != nil {
		return log.Error(fmt.Errorf("could not load legacy storage plugin: %w", err))
	}
	m.pluginConn = conn

	connected := false
	defer func() {
		if !connected {
			conn.Close(ctx)
		}
	}()

	// Cast the plugin connection to a subset of the old protocol from v0.38 that can only read data
	store, ok := conn.GetClient().(crudstore.Store)
	if !ok {
		return log.Error(fmt.Errorf("the interface exposed by the %s plugin was not crudstore.Store", conn))
	}

	m.sourceStore = store
	connected = true
	return nil
}

func (m *Migration) LegacyStoragePluginConfig() pluggable.PluginTypeConfig {
	return pluggable.PluginTypeConfig{
		Interface: plugins.PluginInterface,
		Plugin:    &crudstore.Plugin{},
		GetDefaultPluggable: func(c *config.Config) string {
			// Load the config for the specific storage account named as the source for the migration
			return m.opts.SourceAccount
		},
		GetPluggable: func(c *config.Config, name string) (pluggable.Entry, error) {
			return c.GetStorage(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			// filesystem is the default storage plugin for v0.38
			return "filesystem"
		},
		ProtocolVersion: 1, // protocol version used by porter v0.38
	}
}

func (m *Migration) Close() error {
	m.pluginConn.Close(context.Background())
	return nil
}

func (m *Migration) Migrate(ctx context.Context, currentSchema storage.Schema) (storage.Schema, error) {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("installationSchema", string(currentSchema.Installations)),
		attribute.String("parameterSchema", string(currentSchema.Parameters)),
		attribute.String("credentialSchema", string(currentSchema.Credentials)),
	)
	defer span.EndSpan()

	if err := m.Connect(ctx); err != nil {
		return storage.Schema{}, err
	}

	// Attempt to migrate all data, don't immediately stop when one fails
	// Report how it went at the end
	var err error
	var migrationErr *multierror.Error
	if currentSchema.ShouldMigrateInstallations() {
		span.Info("Installations schema is out-of-date. Migrating...")
		err = m.migrateInstallations(ctx)
		migrationErr = multierror.Append(migrationErr, err)
	} else {
		span.Info("Installations schema is up-to-date")
	}

	if currentSchema.ShouldMigrateCredentialSets() {
		span.Info("Credential Sets schema is out-of-date. Migrating...")
		err = m.migrateCredentialSets(ctx)
		migrationErr = multierror.Append(migrationErr, err)
	} else {
		span.Info("Credential Sets schema is up-to-date")
	}

	if currentSchema.ShouldMigrateParameterSets() {
		span.Info("Parameters schema is out-of-date. Migrating...")
		err = m.migrateParameterSets(ctx)
		migrationErr = multierror.Append(migrationErr, err)
	} else {
		span.Info("Parameter Sets schema is up-to-date")
	}

	if migrationErr.ErrorOrNil() == nil {
		currentSchema, err = WriteSchema(ctx, m.destStore)
		migrationErr = multierror.Append(migrationErr, err)
	}
	return currentSchema, m.migrateInstallations(ctx)
}

func (m *Migration) migrateInstallations(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// Get a list of all the installation names
	names, err := m.listItems("installations", "")
	if err != nil {
		return span.Error(fmt.Errorf("error listing installations from the source account: %w", err))
	}

	var bigErr *multierror.Error
	for _, name := range names {
		if err = m.migrateInstallation(ctx, name); err != nil {
			span.Error(err, attribute.String("installation", name))

			// Keep track of which installations failed but otherwise keep trying to migrate as many as possible
			bigErr = multierror.Append(bigErr, err)
		}
	}

	return bigErr.ErrorOrNil()
}

func (m *Migration) migrateInstallation(ctx context.Context, installationName string) error {
	inst := convertInstallation(installationName)
	inst.Namespace = m.opts.DestinationNamespace

	// Find all claims associated with the installation
	claimIDs, err := m.listItems("claims", installationName)
	if err != nil {
		return err
	}

	for _, claimID := range claimIDs {
		if err = m.migrateClaim(ctx, &inst, claimID); err != nil {
			return err
		}
	}

	err = m.destStore.Insert(ctx, storage.CollectionInstallations, storage.InsertOptions{Documents: []interface{}{inst}})
	return nil
}

func convertInstallation(installationName string) storage.Installation {
	inst := storage.NewInstallation("", installationName)

	// Clear the timestamp and populate them from the claim dates
	inst.Status.Created = time.Time{}
	inst.Status.Modified = time.Time{}

	return inst
}

// migrateClaim migrates the specified claim record into the target database, updating the installation
// status based on all processed claims (such as setting the created date for the installation).
func (m *Migration) migrateClaim(ctx context.Context, inst *storage.Installation, claimID string) error {
	// inst is a ref because migrateClaim will update its status based on the processed claims

	ctx, span := tracing.StartSpan(ctx,
		attribute.String("installation", inst.Name), attribute.String("claimID", claimID))
	defer span.EndSpan()

	data, err := m.sourceStore.Read("claims", claimID)
	if err != nil {
		return span.Error(err)
	}

	run, err := convertClaimToRun(*inst, data)
	if err != nil {
		return span.Error(err)
	}

	// Update the installation status based on the run
	// Use the most early claim timestamp as the creation date of the installation
	if inst.Status.Created.IsZero() || inst.Status.Created.After(run.Created) {
		inst.Status.Created = run.Created
	}

	// Use the most recent claim timestamp as the modified date of the installation
	if inst.Status.Modified.IsZero() || inst.Status.Modified.Before(run.Created) {
		inst.Status.Modified = run.Created
	}

	// Sanitize sensitive values on the source claim
	bun := cnab.ExtendedBundle{Bundle: run.Bundle}
	run.Parameters.Parameters, err = m.sanitizer.CleanParameters(ctx, run.Parameters.Parameters, bun, run.ID)
	if err != nil {
		return span.Error(err)
	}

	// Find all results associated with the run
	resultIDs, err := m.listItems("results", run.ID)
	if err != nil {
		return err
	}

	for _, resultID := range resultIDs {
		if err = m.migrateResult(ctx, inst, run, resultID); err != nil {
			return err
		}
	}

	err = m.destStore.Insert(ctx, storage.CollectionRuns, storage.InsertOptions{Documents: []interface{}{run}})
	if err != nil {
		return span.Error(err)
	}

	return nil
}

func convertClaimToRun(inst storage.Installation, data []byte) (storage.Run, error) {
	var src SourceClaim
	if err := json.Unmarshal(data, &src); err != nil {
		return storage.Run{}, fmt.Errorf("error parsing claim record: %w", err)
	}

	params := make([]secrets.Strategy, 0, len(src.Parameters))
	for k, v := range src.Parameters {
		stringVal, err := cnab.WriteParameterToString(k, v)
		if err != nil {
			return storage.Run{}, err
		}
		params = append(params, storage.ValueStrategy(k, stringVal))
	}

	dest := storage.Run{
		SchemaVersion:   storage.InstallationSchemaVersion,
		ID:              src.ID,
		Created:         src.Created,
		Namespace:       inst.Namespace,
		Installation:    src.Installation,
		Revision:        src.Revision,
		Action:          src.Action,
		Bundle:          src.Bundle,
		BundleReference: src.BundleReference,
		BundleDigest:    "", // We didn't track digest before v1
		Parameters:      storage.NewInternalParameterSet(inst.Namespace, src.ID, params...),
		Custom:          src.Custom,
	}

	return dest, nil
}

func (m *Migration) migrateResult(ctx context.Context, inst *storage.Installation, run storage.Run, resultID string) error {
	// inst is a ref because migrateResult will update the installation status based on the result of the run

	ctx, span := tracing.StartSpan(ctx, attribute.String("resultID", resultID))
	defer span.EndSpan()

	data, err := m.sourceStore.Read("results", resultID)
	if err != nil {
		return span.Error(err)
	}

	result, err := convertResult(run, data)
	if err != nil {
		return span.Error(err)
	}

	err = m.destStore.Insert(ctx, storage.CollectionResults, storage.InsertOptions{Documents: []interface{}{result}})
	if err != nil {
		return span.Error(err)
	}

	// Update the installation status based on the result of previous runs
	inst.ApplyResult(run, result)

	// Find all outputs associated with the result
	outputKeys, err := m.listItems("outputs", resultID)
	if err != nil {
		return err
	}

	for _, outputKey := range outputKeys {
		if err = m.migrateOutput(ctx, run, result, outputKey); err != nil {
			return err
		}
	}

	return nil
}

func convertResult(run storage.Run, data []byte) (storage.Result, error) {
	var src SourceResult
	if err := json.Unmarshal(data, &src); err != nil {
		return storage.Result{}, fmt.Errorf("error parsing result record: %w", err)
	}

	dest := storage.Result{
		SchemaVersion:  run.SchemaVersion,
		ID:             src.ID,
		Created:        src.Created,
		Namespace:      run.Namespace,
		Installation:   run.Installation,
		RunID:          run.ID,
		Message:        src.Message,
		Status:         src.Status,
		OutputMetadata: src.OutputMetadata,
		Custom:         src.Custom,
	}

	return dest, nil
}

func (m *Migration) migrateOutput(ctx context.Context, run storage.Run, result storage.Result, outputKey string) error {
	ctx, span := tracing.StartSpan(ctx, attribute.String("outputKey", outputKey))
	defer span.EndSpan()

	data, err := m.sourceStore.Read("outputs", outputKey)
	if err != nil {
		return span.Error(err)
	}

	output, err := convertOutput(result, outputKey, data)
	if err != nil {
		return span.Error(err)
	}

	// Sanitize sensitive outputs
	bun := cnab.ExtendedBundle{Bundle: run.Bundle}
	output, err = m.sanitizer.CleanOutput(ctx, output, bun)
	if err != nil {
		return span.Error(err)
	}

	err = m.destStore.Insert(ctx, storage.CollectionOutputs, storage.InsertOptions{Documents: []interface{}{output}})
	return span.Error(err)
}

func convertOutput(result storage.Result, outputKey string, data []byte) (storage.Output, error) {
	_, outputName, ok := strings.Cut(outputKey, "-")
	if !ok {
		return storage.Output{}, fmt.Errorf("error converting source output: invalid output key %s", outputKey)
	}

	dest := storage.Output{
		SchemaVersion: result.SchemaVersion,
		Name:          outputName,
		Namespace:     result.Namespace,
		Installation:  result.Installation,
		RunID:         result.RunID,
		ResultID:      result.ID,
		Value:         data,
	}

	return dest, nil
}

func (m *Migration) migrateCredentialSets(ctx context.Context) error {
	panic("not implemented")
}

func (m *Migration) migrateParameterSets(ctx context.Context) error {
	panic("not implemented")
}

// List items in a collection, and safely handles when there are no results
func (m *Migration) listItems(itemType string, group string) ([]string, error) {
	names, err := m.sourceStore.List(itemType, group)
	if err != nil {
		// Check for a sentinel error that was returned from legacy plugins
		// when it couldn't list data because the container for the item or group didn't exist
		// This just means no items were found.
		if strings.Contains(err.Error(), "File does not exist") {
			return nil, nil
		}

		return nil, err
	}

	return names, nil
}
