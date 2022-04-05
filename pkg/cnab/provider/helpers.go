package cnabprovider

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	inmemorysecrets "get.porter.sh/porter/pkg/secrets/plugins/in-memory"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/test"

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
	testSecrets := inmemorysecrets.NewStore()
	testClaims := claims.NewTestClaimProviderFor(tc.TestContext.T, testStorage)
	testCredentials := credentials.NewTestCredentialProviderFor(tc.TestContext.T, testStorage)
	testParameters := parameters.NewTestParameterProviderFor(tc.TestContext.T, testStorage, testSecrets)

	return NewTestRuntimeFor(tc, testClaims, testCredentials, testParameters, testSecrets)
}

func NewTestRuntimeFor(tc *config.TestConfig, testClaims *claims.TestClaimProvider, testCredentials *credentials.TestCredentialProvider, testParameters *parameters.TestParameterProvider, testSecrets secrets.Store) *TestRuntime {
	return &TestRuntime{
		Runtime:         NewRuntime(tc.Config, testClaims, testCredentials, testSecrets),
		TestStorage:     storage.TestStore{},
		TestClaims:      testClaims,
		TestCredentials: testCredentials,
		TestParameters:  testParameters,
		TestConfig:      tc,
	}
}

func (t *TestRuntime) Teardown() error {
	t.TestClaims.Teardown()
	t.TestCredentials.Teardown()
	t.TestParameters.Teardown()
	return nil
}

func (t *TestRuntime) LoadBundle(bundleFile string) (cnab.ExtendedBundle, error) {
	return t.Runtime.LoadBundle(bundleFile)
}

func (t *TestRuntime) LoadTestBundle(bundleFile string) cnab.ExtendedBundle {
	bun, err := cnab.LoadBundle(portercontext.New(), bundleFile)
	require.NoError(t.TestConfig.TestContext.T, err)
	return bun
}

func (t *TestRuntime) Execute(ctx context.Context, args ActionArguments) error {
	if args.Driver == "" {
		args.Driver = debugDriver
	}
	return t.Runtime.Execute(ctx, args)
}

func (t *TestRuntime) MockGetDockerGroupId() {
	// mock retrieving the docker group id on linux
	// This is only called on linux, and we just need to have it return something
	// so that the test doesn't fail
	t.Setenv(test.ExpectedCommandOutputEnv, "docker:x:103")
}
