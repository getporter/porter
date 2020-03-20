package cnabprovider

import (
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
)

type Runtime struct {
	*config.Config
	credentials credentials.CredentialProvider
	claims      claims.ClaimProvider
}

func NewRuntime(c *config.Config, claims claims.ClaimProvider, credentials credentials.CredentialProvider) *Runtime {
	return &Runtime{
		Config:      c,
		claims:      claims,
		credentials: credentials,
	}
}
