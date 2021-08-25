package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/require"
)

const debugDriver = "debug"

var _ CNABProvider = &TestRuntime{}

type TestRuntime struct {
	*Runtime
	TestStorage     storage.TestStore
	TestClaims      *claims.TestClaimProvider
	TestCredentials *credentials.TestCredentialProvider
	TestParameters  *parameters.TestParameterProvider
	TestConfig      *config.TestConfig
}

func NewTestRuntime(t *testing.T) *TestRuntime {
	tc := config.NewTestConfig(t)
	testStorage := storage.NewTestStore(tc.TestContext)
	testClaims := claims.NewTestClaimProviderFor(tc.TestContext.T, testStorage)
	testCredentials := credentials.NewTestCredentialProviderFor(tc.TestContext.T, testStorage)
	testParameters := parameters.NewTestParameterProviderFor(tc.TestContext.T, testStorage)

	return NewTestRuntimeFor(tc, testClaims, testCredentials, testParameters)
}

func NewTestRuntimeFor(tc *config.TestConfig, testClaims *claims.TestClaimProvider, testCredentials *credentials.TestCredentialProvider, testParameters *parameters.TestParameterProvider) *TestRuntime {
	return &TestRuntime{
		TestConfig:      tc,
		TestClaims:      testClaims,
		TestCredentials: testCredentials,
		TestParameters:  testParameters,
		Runtime:         NewRuntime(tc.Config, testClaims, testCredentials),
	}
}

func (t *TestRuntime) Teardown() error {
	t.TestClaims.Teardown()
	t.TestCredentials.Teardown()
	t.TestParameters.Teardown()
	return nil
}

func (t *TestRuntime) LoadBundle(bundleFile string) (bundle.Bundle, error) {
	return t.Runtime.LoadBundle(bundleFile)
}

func (t *TestRuntime) LoadTestBundle(bundleFile string) bundle.Bundle {
	bun, err := cnab.LoadBundle(context.New(), bundleFile)
	require.NoError(t.TestConfig.TestContext.T, err)
	return bun
}

func (t *TestRuntime) Execute(args ActionArguments) error {
	if args.Driver == "" {
		args.Driver = debugDriver
	}
	return t.Runtime.Execute(args)
}
