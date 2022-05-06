package cnabprovider

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/test"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"
)

const debugDriver = "debug"

var _ CNABProvider = &TestRuntime{}

type TestRuntime struct {
	*Runtime
	TestStorage       storage.TestStore
	TestInstallations *storage.TestInstallationProvider
	TestCredentials   *storage.TestCredentialSetProvider
	TestParameters    *storage.TestParameterSetProvider
	TestConfig        *config.TestConfig
}

func NewTestRuntime(t *testing.T) *TestRuntime {
	tc := config.NewTestConfig(t)
	testStorage := storage.NewTestStore(tc)
	testSecrets := secrets.NewTestSecretsProvider()
	testInstallations := storage.NewTestInstallationProviderFor(tc.TestContext.T, testStorage)
	testCredentials := storage.NewTestCredentialProviderFor(tc.TestContext.T, testStorage, testSecrets)
	testParameters := storage.NewTestParameterProviderFor(tc.TestContext.T, testStorage, testSecrets)

	return NewTestRuntimeFor(tc, testInstallations, testCredentials, testParameters, testSecrets)
}

func NewTestRuntimeFor(tc *config.TestConfig, testInstallations *storage.TestInstallationProvider, testCredentials *storage.TestCredentialSetProvider, testParameters *storage.TestParameterSetProvider, testSecrets secrets.Store) *TestRuntime {
	return &TestRuntime{
		Runtime:           NewRuntime(tc.Config, testInstallations, testCredentials, testSecrets, storage.NewSanitizer(testParameters, testSecrets)),
		TestStorage:       storage.TestStore{},
		TestInstallations: testInstallations,
		TestCredentials:   testCredentials,
		TestParameters:    testParameters,
		TestConfig:        tc,
	}
}

func (t *TestRuntime) Close() error {
	var bigErr *multierror.Error

	bigErr = multierror.Append(bigErr, t.TestInstallations.Close())
	bigErr = multierror.Append(bigErr, t.TestCredentials.Close())
	bigErr = multierror.Append(bigErr, t.TestParameters.Close())

	return bigErr.ErrorOrNil()
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
