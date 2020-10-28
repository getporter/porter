package credentials

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	secretplugins "get.porter.sh/porter/pkg/secrets/pluginstore"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/credentials"
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type CredentialsStore = credentials.Store
type SecretsStore = cnabsecrets.Store

var _ CredentialProvider = &CredentialStorage{}

// CredentialStorage provides access to credential sets by instantiating plugins that
// implement CRUD storage.
type CredentialStorage struct {
	*config.Config
	CredentialsStore
	SecretsStore
}

func NewCredentialStorage(storage *storage.Manager) *CredentialStorage {
	credStore := credentials.NewCredentialStore(storage)
	return &CredentialStorage{
		Config:           storage.Config,
		CredentialsStore: credStore,
		SecretsStore:     secrets.NewSecretStore(secretplugins.NewStore(storage.Config)),
	}
}

func (s CredentialStorage) ResolveAll(creds credentials.CredentialSet) (valuesource.Set, error) {
	resolvedCreds := make(valuesource.Set)
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

func (s CredentialStorage) Validate(creds credentials.CredentialSet) error {
	validSources := []string{secrets.SourceSecret, host.SourceValue, host.SourceEnv, host.SourcePath, host.SourceCommand}
	var errors error

	for _, cs := range creds.Credentials {
		valid := false
		for _, validSource := range validSources {
			if cs.Source.Key == validSource {
				valid = true
				break
			}
		}
		if valid == false {
			errors = multierror.Append(errors, fmt.Errorf(
				"%s is not a valid source. Valid sources are: %s",
				cs.Source.Key,
				strings.Join(validSources, ", "),
			))
		}
	}

	return errors
}
