package migrations

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
)

const (
	// CollectionConfig is the collection that stores Porter configuration documents
	// such as the storage schema.
	CollectionConfig = "config"
)

var _ storage.Store = &storage.PluginAdapter{}
var _ storage.Store = &Manager{}

// Manager handles high level functions over Porter's storage systems such as
// migrating data formats.
type Manager struct {
	*config.Config

	// The underlying storage managed by this instance. It
	// shouldn't be used for typed read/access the data, for that storage.InstallationStorageProvider
	// or storage.CredentialSetProvider which works with the Manager.
	store storage.Store

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
func NewManager(c *config.Config, storage storage.Store) *Manager {
	mgr := &Manager{
		Config: c,
		store:  storage,
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

Refer to https://getporter.org/storage-migrate for more information and instructions to back up your data. 
Once your data has been backed up, run the following command to perform the migration:

    porter storage migrate
`, pkg.Version, storage.NewSchema(), m.schema))
		}
		m.initialized = true

		err := storage.EnsureInstallationIndices(ctx, m.store)
		if err != nil {
			return err
		}

		err = storage.EnsureParameterIndices(ctx, m.store)
		if err != nil {
			return err
		}

		err = storage.EnsureCredentialIndices(ctx, m.store)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) Close() error {
	m.store.Close()
	m.initialized = false
	return nil
}

func (m *Manager) Aggregate(ctx context.Context, collection string, opts storage.AggregateOptions, out interface{}) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	return m.store.Aggregate(ctx, collection, opts, out)
}

func (m *Manager) Count(ctx context.Context, collection string, opts storage.CountOptions) (int64, error) {
	if err := m.Connect(ctx); err != nil {
		return 0, err
	}
	return m.store.Count(ctx, collection, opts)
}

func (m *Manager) EnsureIndex(ctx context.Context, opts storage.EnsureIndexOptions) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	return m.store.EnsureIndex(ctx, opts)
}

func (m *Manager) Find(ctx context.Context, collection string, opts storage.FindOptions, out interface{}) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	return m.store.Find(ctx, collection, opts, out)
}

func (m *Manager) FindOne(ctx context.Context, collection string, opts storage.FindOptions, out interface{}) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	return m.store.FindOne(ctx, collection, opts, out)
}

func (m *Manager) Get(ctx context.Context, collection string, opts storage.GetOptions, out interface{}) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	return m.store.Get(ctx, collection, opts, out)
}

func (m *Manager) Insert(ctx context.Context, collection string, opts storage.InsertOptions) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	return m.store.Insert(ctx, collection, opts)
}

func (m *Manager) Patch(ctx context.Context, collection string, opts storage.PatchOptions) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	return m.store.Patch(ctx, collection, opts)
}

func (m *Manager) Remove(ctx context.Context, collection string, opts storage.RemoveOptions) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	return m.store.Remove(ctx, collection, opts)
}

func (m *Manager) Update(ctx context.Context, collection string, opts storage.UpdateOptions) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	return m.store.Update(ctx, collection, opts)
}

// loadSchema parses the schema file at the root of PORTER_HOME. This file (when present) contains
// a list of the current version of each of Porter's storage systems.
func (m *Manager) loadSchema(ctx context.Context) error {
	var schema storage.Schema
	err := m.store.Get(ctx, CollectionConfig, storage.GetOptions{ID: "schema"}, &schema)
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

// Migrate executes a migration on any/all of Porter's storage sub-systems.
// You must call Initialize before calling Migrate.
func (m *Manager) Migrate(ctx context.Context, opts storage.MigrateOptions) error {
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
	return nil
}

// When there is no schema, and no existing storage data, create an initial
// schema file and allow the operation to continue. Don't require a
// migration.
func (m *Manager) initEmptyPorterHome(ctx context.Context) (bool, error) {
	if m.schema != (storage.Schema{}) {
		return false, nil
	}

	itemCheck := func(itemType string) (bool, error) {
		itemCount, err := m.store.Count(ctx, itemType, storage.CountOptions{})
		if err != nil {
			return false, fmt.Errorf("error checking for existing %s when checking if PORTER_HOME is new: %w", itemType, err)
		}

		return itemCount > 0, nil
	}

	hasInstallations, err := itemCheck(storage.CollectionInstallations)
	if hasInstallations || err != nil {
		return false, err
	}

	hasCredentials, err := itemCheck(storage.CollectionCredentials)
	if hasCredentials || err != nil {
		return false, err
	}

	hasParameters, err := itemCheck(storage.CollectionParameters)
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
func WriteSchema(ctx context.Context, store storage.Store) (storage.Schema, error) {
	schema := storage.NewSchema()

	err := store.Update(ctx, CollectionConfig, storage.UpdateOptions{Document: schema, Upsert: true})
	if err != nil {
		return storage.Schema{}, fmt.Errorf("Unable to save storage schema file to the database: %w", err)
	}

	return schema, nil
}
