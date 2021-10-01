package migrations

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/schema"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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
	// shouldn't be used for typed read/access the data, for that use the ClaimsProvider
	// or CredentialsProvider which works with the Storage.Manager.
	store storage.Store

	// initialized specifies if we have loaded the schema document.
	initialized bool

	// schema document that defines the current version of each storage system.
	// We use this to detect when we are out-of-date and require a migration.
	schema storage.Schema

	// Allow the schema to be out-of-date, defaults to false. Prevents
	// connections to underlying storage when the schema is out-of-date
	allowOutOfDateSchema bool
}

// NewManager creates a storage manager for a backing datastore.
func NewManager(c *config.Config, storage storage.Store) *Manager {
	mgr := &Manager{
		Config: c,
		store:  storage,
	}

	return mgr
}

func (m *Manager) Connect() error {
	if !m.initialized {
		if err := m.store.Connect(); err != nil {
			return err
		}

		if err := m.loadSchema(); err != nil {
			return err
		}

		if !m.allowOutOfDateSchema && m.MigrationRequired() {
			m.Close()
			return errors.New(`The schema of Porter's data is in an older format than supported by this version of Porter. 
Refer to https://porter.sh/storage-migrate for more information and instructions to back up your data. 
Once your data has been backed up, run the following command to perform the migration:

    porter storage migrate
`)
		}
		m.initialized = true

		cs := claims.NewClaimStore(m.store)
		err := cs.Initialize()
		if err != nil {
			return err
		}

		paramStore := parameters.NewParameterStore(m.store, nil)
		err = paramStore.Initialize()
		if err != nil {
			return err
		}

		credStore := credentials.NewCredentialStore(m.store, nil)
		err = credStore.Initialize()
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

func (m *Manager) GetDataStore() storage.Store {
	return m.store
}

func (m *Manager) Aggregate(collection string, opts storage.AggregateOptions, out interface{}) error {
	if err := m.Connect(); err != nil {
		return err
	}
	return m.store.Aggregate(collection, opts, out)
}

func (m *Manager) Count(collection string, opts storage.CountOptions) (int64, error) {
	if err := m.Connect(); err != nil {
		return 0, err
	}
	return m.store.Count(collection, opts)
}

func (m *Manager) EnsureIndex(opts storage.EnsureIndexOptions) error {
	if err := m.Connect(); err != nil {
		return err
	}
	return m.store.EnsureIndex(opts)
}

func (m *Manager) Find(collection string, opts storage.FindOptions, out interface{}) error {
	if err := m.Connect(); err != nil {
		return err
	}
	return m.store.Find(collection, opts, out)
}

func (m *Manager) FindOne(collection string, opts storage.FindOptions, out interface{}) error {
	if err := m.Connect(); err != nil {
		return err
	}
	return m.store.FindOne(collection, opts, out)
}

func (m *Manager) Get(collection string, opts storage.GetOptions, out interface{}) error {
	if err := m.Connect(); err != nil {
		return err
	}
	return m.store.Get(collection, opts, out)
}

func (m *Manager) Insert(collection string, opts storage.InsertOptions) error {
	if err := m.Connect(); err != nil {
		return err
	}
	return m.store.Insert(collection, opts)
}

func (m *Manager) Patch(collection string, opts storage.PatchOptions) error {
	if err := m.Connect(); err != nil {
		return err
	}
	return m.store.Patch(collection, opts)
}

func (m *Manager) Remove(collection string, opts storage.RemoveOptions) error {
	if err := m.Connect(); err != nil {
		return err
	}
	return m.store.Remove(collection, opts)
}

func (m *Manager) Update(collection string, opts storage.UpdateOptions) error {
	if err := m.Connect(); err != nil {
		return err
	}
	return m.store.Update(collection, opts)
}

// loadSchema parses the schema file at the root of PORTER_HOME. This file (when present) contains
// a list of the current version of each of Porter's storage systems.
func (m *Manager) loadSchema() error {
	var schema storage.Schema
	err := m.store.Get(CollectionConfig, storage.GetOptions{ID: "schema"}, &schema)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound{}) {
			emptyHome, err := m.initEmptyPorterHome()
			if emptyHome {
				// Return any errors from creating a schema document in an empty porter home directory
				return err
			} else {
				// When we don't have an empty home directory, and we can't find the schema
				// document, we need to do a migration
				return nil
			}
		}
		return errors.Wrap(err, "could not read storage schema document")
	}

	m.schema = schema

	return errors.Wrap(err, "could not parse storage schema document")
}

// MigrationRequired determines if a migration of Porter's storage system is necessary.
func (m *Manager) MigrationRequired() bool {
	return m.ShouldMigrateClaims() || m.ShouldMigrateCredentials() || m.ShouldMigrateParameters()
}

// Migrate executes a migration on any/all of Porter's storage sub-systems.
func (m *Manager) Migrate() (string, error) {
	m.reset()

	// Let us call connect and not have it kick us out because the schema is out-of-date
	m.allowOutOfDateSchema = true
	defer func() {
		m.allowOutOfDateSchema = false
	}()

	// Reuse the same connection for the entire migration
	err := m.Connect()
	if err != nil {
		return "", err
	}
	defer m.Close()

	home, err := m.GetHomeDir()
	if err != nil {
		return "", err
	}

	logfilePath := filepath.Join(home, fmt.Sprintf("%s-migrate.log", time.Now().Format("20060102150405")))
	logfile, err := m.FileSystem.Create(logfilePath)
	if err != nil {
		return "", errors.Wrapf(err, "error creating logfile for migration at %s", logfilePath)
	}
	defer logfile.Close()
	w := io.MultiWriter(m.Err, logfile)

	var migrationErr *multierror.Error
	if m.ShouldMigrateClaims() {
		fmt.Fprintf(w, "Installations schema is out-of-date (want: %s got: %s)\n", claims.SchemaVersion, m.schema.Installations)
		err = m.migrateClaims()
		migrationErr = multierror.Append(migrationErr, err)
	} else {
		fmt.Fprintln(w, "Installations schema is up-to-date")
	}

	if m.ShouldMigrateCredentials() {
		fmt.Fprintf(w, "Credentials schema is out-of-date (want: %s got: %s)\n", credentials.SchemaVersion, m.schema.Credentials)
		err = m.migrateCredentials(w)
		migrationErr = multierror.Append(migrationErr, err)
	} else {
		fmt.Fprintln(w, "Credentials schema is up-to-date")
	}

	if m.ShouldMigrateParameters() {
		fmt.Fprintf(w, "Parameters schema is out-of-date (want: %s got: %s)\n", parameters.SchemaVersion, m.schema.Parameters)
		err = m.migrateParameters(w)
		migrationErr = multierror.Append(migrationErr, err)
	} else {
		fmt.Fprintln(w, "Parameters schema is up-to-date")
	}

	if migrationErr.ErrorOrNil() == nil {
		err = m.WriteSchema()
		migrationErr = multierror.Append(migrationErr, err)
	}

	return logfilePath, migrationErr.ErrorOrNil()
}

// When there is no schema, and no existing storage data, create an initial
// schema file and allow the operation to continue. Don't require a
// migration.
func (m *Manager) initEmptyPorterHome() (bool, error) {
	if m.schema != (storage.Schema{}) {
		return false, nil
	}

	itemCheck := func(itemType string) (bool, error) {
		itemCount, err := m.store.Count(itemType, storage.CountOptions{})
		if err != nil {
			return false, errors.Wrapf(err, "error checking for existing %s when checking if PORTER_HOME is new", itemType)
		}

		return itemCount > 0, nil
	}

	hasInstallations, err := itemCheck(claims.CollectionInstallations)
	if hasInstallations || err != nil {
		return false, err
	}

	hasCredentials, err := itemCheck(credentials.CollectionCredentials)
	if hasCredentials || err != nil {
		return false, err
	}

	hasParameters, err := itemCheck(parameters.CollectionParameters)
	if hasParameters || err != nil {
		return false, err
	}

	return true, m.WriteSchema()
}

// ShouldMigrateClaims determines if the claims storage system requires a migration.
func (m *Manager) ShouldMigrateClaims() bool {
	return m.schema.Installations != claims.SchemaVersion
}

func (m *Manager) migrateClaims() error {
	return nil
}

// reset allows us to relook at our schema.json even after it has been read.
func (m *Manager) reset() {
	m.schema = storage.Schema{}
	m.initialized = false
}

// WriteSchema updates the schema with the most recent version then writes it to disk.
func (m *Manager) WriteSchema() error {
	m.schema = NewSchema()

	err := m.store.Update(CollectionConfig, storage.UpdateOptions{Document: m.schema, Upsert: true})
	if err != nil {
		return errors.Wrap(err, "Unable to save storage schema file")
	}

	return nil
}

// NewSchema creates a new schema document for the current version of the CNAB spec used by Porter.
func NewSchema() storage.Schema {
	return storage.NewSchema(
		claims.SchemaVersion,
		credentials.SchemaVersion,
		parameters.SchemaVersion)
}

// ShouldMigrateCredentials determines if the credentials storage system requires a migration.
func (m *Manager) ShouldMigrateCredentials() bool {
	return m.schema.Credentials != credentials.SchemaVersion
}

func (m *Manager) migrateCredentials(w io.Writer) error {
	return nil
}

// ShouldMigrateParameters determines if the parameter set documents requires a migration.
func (m *Manager) ShouldMigrateParameters() bool {
	return m.schema.Parameters != parameters.SchemaVersion
}

func (m *Manager) migrateParameters(w io.Writer) error {
	return nil
}

// getSchemaVersion attempts to read the schemaVersion stamped on a document.
func getSchemaVersion(data []byte) string {
	var peek struct {
		SchemaVersion schema.Version `json:"schemaVersion"`
	}

	err := json.Unmarshal(data, &peek)
	if err != nil {
		return "unknown"
	}

	return string(peek.SchemaVersion)
}
