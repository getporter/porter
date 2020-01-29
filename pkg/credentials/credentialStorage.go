package credentials

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	secretplugins "get.porter.sh/porter/pkg/secrets/pluginstore"
	crudplugins "get.porter.sh/porter/pkg/storage/pluginstore"
	"github.com/cnabio/cnab-go/credentials"
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type CredentialsStore = credentials.Store
type SecretsStore = cnabsecrets.Store

var _ CredentialProvider = &CredentialStorage{}

// Credentials provides access to credential sets by instantiating plugins that
// implement CRUD storage.
type CredentialStorage struct {
	*config.Config
	*CredentialsStore
	SecretsStore
}

func NewCredentialStorage(c *config.Config, storagePlugin *crudplugins.Store) *CredentialStorage {
	migration := newMigrateCredentialsWrapper(c, storagePlugin)
	credStore := credentials.NewCredentialStore(migration)
	return &CredentialStorage{
		Config:           c,
		CredentialsStore: &credStore,
		SecretsStore:     secrets.NewSecretStore(secretplugins.NewStore(c)),
	}
}

func (s CredentialStorage) ResolveAll(creds credentials.CredentialSet) (credentials.Set, error) {
	resolvedCreds := make(credentials.Set)
	var resolveErrors error

	for _, cred := range creds.Credentials {
		value, err := s.Resolve(cred.Source.Key, cred.Source.Value)
		if err != nil {
			resolveErrors = multierror.Append(resolveErrors, errors.Wrapf(err, "unable to resolve credential %s.%s from %s %s", creds.Name, cred.Name, cred.Source.Key, cred.Source.Value))
		}

		resolvedCreds[cred.Name] = value
	}

	return resolvedCreds, resolveErrors
}
