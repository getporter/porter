package parameters

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	secretplugins "get.porter.sh/porter/pkg/secrets/pluginstore"
	crudplugins "get.porter.sh/porter/pkg/storage/pluginstore"
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// TODO: clone of credentialStorage.go in credentials pkg
// Can generalize/share/DRY out?

type ParametersStore = Store
type SecretsStore = cnabsecrets.Store

var _ ParameterProvider = &ParameterStorage{}

// ParameterStorage provides access to parameter sets by instantiating plugins that
// implement CRUD storage.
type ParameterStorage struct {
	*config.Config
	*ParametersStore
	SecretsStore
}

func NewParameterStorage(c *config.Config, storagePlugin *crudplugins.Store) *ParameterStorage {
	paramStore := NewParameterStore(storagePlugin)
	return &ParameterStorage{
		Config:          c,
		ParametersStore: &paramStore,
		SecretsStore:    secrets.NewSecretStore(secretplugins.NewStore(c)),
	}
}

func (s ParameterStorage) ResolveAll(params ParameterSet) (valuesource.Set, error) {
	resolvedParams := make(valuesource.Set)
	var resolveErrors error

	for _, param := range params.Parameters {
		value, err := s.Resolve(param.Source.Key, param.Source.Value)
		if err != nil {
			resolveErrors = multierror.Append(resolveErrors, errors.Wrapf(err, "unable to resolve parameter %s.%s from %s %s", params.Name, param.Name, param.Source.Key, param.Source.Value))
		}

		resolvedParams[param.Name] = value
	}

	return resolvedParams, resolveErrors
}

func (s ParameterStorage) Validate(params ParameterSet) error {
	validSources := []string{secrets.SourceSecret, host.SourceValue, host.SourceEnv, host.SourcePath, host.SourceCommand}
	var errors error

	for _, cs := range params.Parameters {
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
