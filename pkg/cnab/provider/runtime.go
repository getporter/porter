package cnabprovider

import (
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab/extensions"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/parameters"
)

var _ CNABProvider = &Runtime{}

type Runtime struct {
	*config.Config
	credentials credentials.CredentialProvider
	parameters  parameters.ParameterProvider
	claims      claims.ClaimProvider
	Extensions  extensions.ProcessedExtensions
}

func NewRuntime(c *config.Config, claims claims.ClaimProvider, credentials credentials.CredentialProvider, parameters parameters.ParameterProvider) *Runtime {
	return &Runtime{
		Config:      c,
		claims:      claims,
		credentials: credentials,
		parameters:  parameters,
		Extensions:  extensions.ProcessedExtensions{},
	}
}
