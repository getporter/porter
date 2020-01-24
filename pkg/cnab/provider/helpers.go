package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
)

type TestRuntime struct {
	*Runtime
	TestConfig *config.TestConfig
}

func NewTestRuntime(t *testing.T) *TestRuntime {
	c := config.NewTestConfig(t)
	claimStorage := claims.NewTestClaimProvider()
	credentialStorage := credentials.NewTestCredentialProvider(t, c)

	return &TestRuntime{
		TestConfig: c,
		Runtime:    NewRuntime(c.Config, claimStorage, credentialStorage),
	}
}
