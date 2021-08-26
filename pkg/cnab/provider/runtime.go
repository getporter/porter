package cnabprovider

import (
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/parameters"
	"github.com/pkg/errors"
)

var _ CNABProvider = &Runtime{}

type Runtime struct {
	*config.Config
	credentials credentials.Provider
	parameters  parameters.Provider
	claims      claims.Provider
	Extensions  cnab.ProcessedExtensions
}

func NewRuntime(c *config.Config, claims claims.Provider, credentials credentials.Provider) *Runtime {
	return &Runtime{
		Config:      c,
		claims:      claims,
		credentials: credentials,
		Extensions:  cnab.ProcessedExtensions{},
	}
}

func (r *Runtime) ProcessRequiredExtensions(b cnab.ExtendedBundle) error {
	exts, err := b.ProcessRequiredExtensions()
	if err != nil {
		return errors.Wrap(err, "unable to process required extensions")
	}
	r.Extensions = exts
	return nil
}
