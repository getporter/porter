package migrations

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"get.porter.sh/porter/pkg/tracing"
)

const (
	// CollectionConfig is the collection that stores Porter configuration documents
	// such as the storage schema.
	CollectionConfig = "config"
)

var _ plugins.StorageBackend = &Manager{}

// Manager handles high level functions over Porter's storage systems such as
// migrating data formats.
type Manager struct {
	*config.Config

	// The underlying storage managed by this instance. It
	// shouldn't be used for typed read/access the data, for that storage.InstallationStorageProvider
	// or storage.CredentialSetProvider which works with the Manager.
	store storage.BackendClient

	// initialized specifies if we have loaded the schema document.
	initialized bool

	// schema document that defines the current version of each storage system.
	// We use this to detect when we are out-of-date and require a migration.
	schema storage.Schema

	// Allow the schema to be out-of-date, defaults to false. Prevents
	// connections to underlying storage when the schema is out-of-date
	allowOutOfDateSchema bool

	// Cleans sensitive data so that we don't store it in our database
	// This is set by Initialize not the constructor due to a bit of a circular reference between Manager, ParameterStore, SecretStore and Sanitizer.
	// Improvements could most definitely be made here!
	sanitizer *storage.Sanitizer
}

// NewManager creates a storage manager for a backing datastore.
func NewManager(c *config.Config, backend plugins.StorageBackend) *Manager {
	mgr := &Manager{
		Config: c,
		store:  storage.NewBackendClient(backend),
		// We can't set sanitizer here yet, it is set in Initialize
	}

	return mgr
}

// Initialize configures the storage manager with additional configuration that wasn't available
// when the manager instance was created.
func (m *Manager) Initialize(sanitizer *storage.Sanitizer) {
	m.sanitizer = sanitizer
}

// Connect initializes storage manager for use.
// The manager itself is responsible for ensuring it was called.
// Close is called automatically when the manager is used by Porter.
func (m *Manager) Connect(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if !m.initialized {
		hash, hashErr := connectionHash(m.Config)
		useCache := !m.allowOutOfDateSchema && hashErr == nil

		if useCache {
			if entry, ok := m.loadInitCacheEntry(ctx, hash); ok {
				span.Debug("Using cached storage init check")
				m.schema = entry.Schema
				m.initialized = true
				return nil
			}
		}

		// Make sure that the backend has the tables, indexes, etc. that Porter relies on
		// before reading anything from it. A new database has none of them yet.
		// A verified connection is cached (see above), same as the schema check, so
		// the backend isn't asked to migrate on every command.
		if _, err := m.store.Backend().Migrate(ctx, &protov4.MigrateRequest{}); err != nil {
			return span.Error(fmt.Errorf("could not prepare the storage backend: %w", err))
		}

		span.Debug("Checking database schema")

		if err := m.loadSchema(ctx); err != nil {
			return err
		}

		if !m.allowOutOfDateSchema && m.schema.IsOutOfDate() {
			m.Close()
			return span.Error(fmt.Errorf(`The schema of Porter's data is in an older format than supported by this version of Porter. 

Porter %s uses the following database schema:

%#v

Your database schema is:

%#v

Refer to https://porter.sh/storage-migrate for more information and instructions to back up your data. 
Once your data has been backed up, run the following command to perform the migration:

    porter storage migrate
`, pkg.Version, storage.NewSchema(), m.schema))
		}

		// Only mark the manager as initialized once everything succeeded,
		// so that a failure is retried on the next call.
		m.initialized = true

		if useCache {
			m.saveInitCacheEntry(ctx, hash, m.schema)
		}
	}

	return nil
}

func (m *Manager) Close() error {
	m.store.Close()
	m.initialized = false
	return nil
}

// The methods below implement plugins.StorageBackend. They connect the manager,
// which verifies the schema of the data, before calling the underlying backend.

func (m *Manager) Info(ctx context.Context, req *protov4.InfoRequest) (*protov4.InfoResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().Info(ctx, req)
}

func (m *Manager) Migrate(ctx context.Context, req *protov4.MigrateRequest) (*protov4.MigrateResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().Migrate(ctx, req)
}

func (m *Manager) Put(ctx context.Context, req *protov4.PutRequest) (*protov4.PutResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().Put(ctx, req)
}

func (m *Manager) Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().Get(ctx, req)
}

func (m *Manager) List(ctx context.Context, req *protov4.ListRequest) (*protov4.ListResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().List(ctx, req)
}

func (m *Manager) Remove(ctx context.Context, req *protov4.RemoveRequest) (*protov4.RemoveResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().Remove(ctx, req)
}

func (m *Manager) RemoveInstallation(ctx context.Context, req *protov4.RemoveInstallationRequest) (*protov4.RemoveInstallationResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().RemoveInstallation(ctx, req)
}

func (m *Manager) ListRuns(ctx context.Context, req *protov4.ListRunsRequest) (*protov4.ListRunsResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().ListRuns(ctx, req)
}

func (m *Manager) GetLastRun(ctx context.Context, req *protov4.InstallationRef) (*protov4.GetResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().GetLastRun(ctx, req)
}

func (m *Manager) GetActiveRuns(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().GetActiveRuns(ctx, req)
}

func (m *Manager) ListResults(ctx context.Context, req *protov4.ListResultsRequest) (*protov4.ListResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().ListResults(ctx, req)
}

func (m *Manager) InsertOutputs(ctx context.Context, req *protov4.InsertOutputsRequest) (*protov4.InsertOutputsResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().InsertOutputs(ctx, req)
}

func (m *Manager) ListOutputs(ctx context.Context, req *protov4.ListOutputsRequest) (*protov4.ListResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().ListOutputs(ctx, req)
}

func (m *Manager) GetOutput(ctx context.Context, req *protov4.GetOutputRequest) (*protov4.GetResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().GetOutput(ctx, req)
}

func (m *Manager) GetLastOutputs(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().GetLastOutputs(ctx, req)
}

func (m *Manager) FindInstallationsReferencing(ctx context.Context, req *protov4.FindReferencingRequest) (*protov4.ListResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().FindInstallationsReferencing(ctx, req)
}

func (m *Manager) FindDependencyCandidates(ctx context.Context, req *protov4.FindDependencyCandidatesRequest) (*protov4.ListResponse, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	return m.store.Backend().FindDependencyCandidates(ctx, req)
}

// loadSchema parses the schema file at the root of PORTER_HOME. This file (when present) contains
// a list of the current version of each of Porter's storage systems.
func (m *Manager) loadSchema(ctx context.Context) error {
	schema, err := m.store.GetSchema(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound{}) {
			emptyHome, err := m.initEmptyPorterHome(ctx)
			if emptyHome {
				// Return any errors from creating a schema document in an empty porter home directory
				return err
			} else {
				// When we don't have an empty home directory, and we can't find the schema
				// document, we need to do a migration
				return nil
			}
		}
		return fmt.Errorf("could not read storage schema document: %w", err)
	}

	m.schema = schema

	if err != nil {
		return fmt.Errorf("could not parse storage schema document: %w", err)
	}

	return nil
}

// MigrateData executes a migration on any/all of Porter's storage sub-systems.
// You must call Initialize before calling MigrateData.
func (m *Manager) MigrateData(ctx context.Context, opts storage.MigrateOptions) error {
	if m.sanitizer == nil {
		return fmt.Errorf("cannot call storage.Manager.Migrate before calling Initialize and passing a storage.Sanitizer")
	}

	m.reset()

	// Let us call connect and not have it kick us out because the schema is out-of-date
	m.allowOutOfDateSchema = true
	defer func() {
		m.allowOutOfDateSchema = false
	}()

	// Reuse the same connection for the entire migration
	err := m.Connect(ctx)
	if err != nil {
		return err
	}
	defer m.Close()

	migration := NewMigration(m.Config, opts, m.store, m.sanitizer)
	defer migration.Close()

	newSchema, err := migration.Migrate(ctx)
	if err != nil {
		return err
	}

	m.schema = newSchema

	if hash, hashErr := connectionHash(m.Config); hashErr == nil {
		m.saveInitCacheEntry(ctx, hash, m.schema)
	}

	return nil
}

// When there is no schema, and no existing storage data, create an initial
// schema file and allow the operation to continue. Don't require a
// migration.
func (m *Manager) initEmptyPorterHome(ctx context.Context) (bool, error) {
	if m.schema != (storage.Schema{}) {
		return false, nil
	}

	itemCheck := func(kind protov4.Kind) (bool, error) {
		found, err := m.store.HasDocuments(ctx, kind)
		if err != nil {
			return false, fmt.Errorf("error checking for existing %s when checking if PORTER_HOME is new: %w", kind, err)
		}

		return found, nil
	}

	hasInstallations, err := itemCheck(protov4.Kind_KIND_INSTALLATION)
	if hasInstallations || err != nil {
		return false, err
	}

	hasCredentials, err := itemCheck(protov4.Kind_KIND_CREDENTIAL_SET)
	if hasCredentials || err != nil {
		return false, err
	}

	hasParameters, err := itemCheck(protov4.Kind_KIND_PARAMETER_SET)
	if hasParameters || err != nil {
		return false, err
	}

	return true, m.WriteSchema(ctx)
}

// reset allows us to relook at our schema.json even after it has been read.
func (m *Manager) reset() {
	m.schema = storage.Schema{}
	m.initialized = false
}

// WriteSchema updates the database to indicate that it conforms with the current database schema.
func (m *Manager) WriteSchema(ctx context.Context) error {
	var err error
	m.schema, err = WriteSchema(ctx, m.store)
	return err
}

// WriteSchema updates the database to indicate that it conforms with the current database schema.
func WriteSchema(ctx context.Context, store storage.BackendClient) (storage.Schema, error) {
	schema := storage.NewSchema()

	err := store.PutSchema(ctx, schema)
	if err != nil {
		return storage.Schema{}, fmt.Errorf("Unable to save storage schema file to the database: %w", err)
	}

	return schema, nil
}
