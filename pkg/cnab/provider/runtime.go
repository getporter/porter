package cnabprovider

import (
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
)

var _ CNABProvider = &Runtime{}

type Runtime struct {
	*config.Config
	credentials   storage.CredentialSetProvider
	parameters    storage.ParameterSetProvider
	secrets       secrets.Store
	installations storage.InstallationProvider
	sanitizer     *storage.Sanitizer
	Extensions    cnab.ProcessedExtensions
}

func NewRuntime(c *config.Config, installations storage.InstallationProvider, credentials storage.CredentialSetProvider, secrets secrets.Store, sanitizer *storage.Sanitizer) *Runtime {
	return &Runtime{
		Config:        c,
		installations: installations,
		credentials:   credentials,
		secrets:       secrets,
		sanitizer:     sanitizer,
		Extensions:    cnab.ProcessedExtensions{},
	}
}

func (r *Runtime) ProcessRequiredExtensions(b cnab.ExtendedBundle) error {
	exts, err := b.ProcessRequiredExtensions()
	if err != nil {
		return fmt.Errorf("unable to process required extensions: %w", err)
	}
	r.Extensions = exts
	return nil
}
