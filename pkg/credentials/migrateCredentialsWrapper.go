package credentials

import (
	"path/filepath"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

// Wrapper lets us shim in the migration of the credentials from yaml to json
// transparently for a period of time transparently.
// When we are ready to remove this we can remove the wrapper and go back to directly
// giving the CredentialStore the wrappedStore.
type migrateCredentialsWrapper struct {
	*config.Config
	*crud.BackingStore

	wrappedStore crud.Store
}

func newMigrateCredentialsWrapper(c *config.Config, wrappedStore crud.Store) *migrateCredentialsWrapper {
	return &migrateCredentialsWrapper{
		Config:       c,
		wrappedStore: wrappedStore,
	}
}

func (w *migrateCredentialsWrapper) Connect() error {
	if w.BackingStore != nil {
		return nil
	}

	home, err := w.GetHomeDir()
	if err != nil {
		return errors.Wrap(err, "could not migrate credentials directory")
	}

	credsDir := filepath.Join(home, "credentials")

	migration := NewCredentialsMigration(w.Context)
	err = migration.Migrate(credsDir)
	if err != nil {
		errors.Wrap(err, "conversion of the credentials directory failed")
	}

	w.BackingStore = crud.NewBackingStore(w.wrappedStore)
	return nil
}
