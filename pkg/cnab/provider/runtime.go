package cnabprovider

import (
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab/extensions"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
)

var _ CNABProvider = &Runtime{}

type Runtime struct {
	*config.Config
	credentials credentials.CredentialProvider
	claims      claims.ClaimProvider
	Extensions  extensions.ProcessedExtensions
}

func NewRuntime(c *config.Config, claims claims.ClaimProvider, credentials credentials.CredentialProvider) *Runtime {
	return &Runtime{
		Config:      c,
		claims:      claims,
		credentials: credentials,
		Extensions:  extensions.ProcessedExtensions{},
	}
}
