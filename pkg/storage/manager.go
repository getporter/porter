package storage

import (
	"encoding/json"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

// Manager handles high level functions over Porter's storage systems such as
// migrating data formats.
type Manager struct {
	*config.Config

	// BackingStore is the underlying storage managed by this instance. It
	// shouldn't be used for typed read/access the data, for that use the ClaimsProvider
	// or CredentialsProvider which works with the Storage.Manager.
	*crud.BackingStore

	// connMgr is responsible for providing a consolidated HandleConnect
	// implementation that merges our Connect/Close with those of the datastore.
	connMgr *crud.BackingStore

	// schemaLoaded specifies if we have loaded the schema document.
	schemaLoaded bool

	// schema document that defines the current version of each storage system.
	// We use this to detect when we are out-of-date and require a migration.
	schema Schema

	// Allow the schema to be out-of-date, defaults to false. Prevents
	// connections to underlying storage when the schema is out-of-date
	allowOutOfDateSchema bool
}

// NewManager creates a storage manager for a backing datastore.
func NewManager(c *config.Config, datastore crud.Store) *Manager {
	mgr := &Manager{
		Config:       c,
		BackingStore: crud.NewBackingStore(datastore),
	}

	mgr.connMgr = crud.NewBackingStore(mgr)

	return mgr
}

func (m *Manager) Connect() error {
	err := m.BackingStore.Connect()
	if err != nil {
		return err
	}

	if !m.schemaLoaded {
		if err := m.loadSchema(); err != nil {
			return err
		}

		if !m.allowOutOfDateSchema && m.MigrationRequired() {
			m.Close()
			return errors.New(`The schema of Porter's data is in an older format than supported by this version of Porter. 
Refer to https://porter.sh/migrate for more information and instructions to back up your data. 
Once your data has been backed up, run the following command to perform the migration:

    porter storage migrate
`)
		}
		m.schemaLoaded = true
	}

	return nil
}

func (m *Manager) Close() error {
	return m.BackingStore.Close()
}

func (m *Manager) HandleConnect() (func() error, error) {
	return m.connMgr.HandleConnect()
}

// loadSchema parses the schema file at the root of PORTER_HOME. This file (when present) contains
// a list of the current version of each of Porter's storage systems.
func (p *Manager) loadSchema() error {
	b, err := p.Store.Read("", "schema")
	if err != nil {
		// If schema cannot be found, we'll do a migration
		if !strings.Contains(err.Error(), crud.ErrRecordDoesNotExist.Error()) {
			return errors.Wrap(err, "could not read storage schema document")
		}
	} else {
		err = json.Unmarshal(b, &p.schema)
		if err != nil {
			return errors.Wrap(err, "could not parse storage schema document")
		}
	}

	return nil
}

// MigrationRequired determines if a migration of Porter's storage system is necessary.
func (p *Manager) MigrationRequired() bool {
	// TODO (carolynvs): Include credentials/parameters
	return p.ShouldMigrateClaims()
}

// Migrate executes a migration on any/all of Porter's storage sub-systems.
func (p *Manager) Migrate() error {
	// write the logs for the migration to a file
	return errors.New("not implemented")
}

// ShouldMigrateClaims determines if the claims storage system requires a migration.
func (p *Manager) ShouldMigrateClaims() bool {
	return string(p.schema.Claims) != claim.CNABSpecVersion
}

// ShouldMigrateCredentials determines if the credentials storage system requires a migration.
func (p *Manager) ShouldMigrateCredentials() bool {
	// TODO (carolynvs): This isn't in cnab-go and param sets aren't in the spec...
	return false
}
