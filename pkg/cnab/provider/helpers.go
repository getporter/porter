package cnabprovider

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/storage"
	"testing"
)

type TestRuntime struct {
	*Runtime
	TestConfig *config.TestConfig
}

func NewTestRuntime(t *testing.T) *TestRuntime {
	c := config.NewTestConfig(t)
	claimStorage := storage.NewTestClaimProvider()
	credentialStorage := credentials.NewTestCredentialProvider(t, c)

	return &TestRuntime{
		TestConfig:   c,
		Runtime: NewRuntime(c.Config, claimStorage, credentialStorage),
	}
}
