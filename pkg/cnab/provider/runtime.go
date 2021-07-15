package cnabprovider

import (
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab/extensions"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/parameters"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

var _ CNABProvider = &Runtime{}

type Runtime struct {
	*config.Config
	credentials credentials.Provider
	parameters  parameters.Provider
	claims      claims.Provider
	Extensions  extensions.ProcessedExtensions
}

func NewRuntime(c *config.Config, claims claims.Provider, credentials credentials.Provider) *Runtime {
	return &Runtime{
		Config:      c,
		claims:      claims,
		credentials: credentials,
		Extensions:  extensions.ProcessedExtensions{},
	}
}

func (r *Runtime) ProcessRequiredExtensions(b bundle.Bundle) error {
	exts, err := extensions.ProcessRequiredExtensions(b)
	if err != nil {
		return errors.Wrap(err, "unable to process required extensions")
	}
	r.Extensions = exts
	return nil
}
