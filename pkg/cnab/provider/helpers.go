package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/parameters"
	"github.com/cnabio/cnab-go/bundle"
)

const debugDriver = "debug"

var _ CNABProvider = &TestRuntime{}

type TestRuntime struct {
	*Runtime
	TestCredentials credentials.TestCredentialProvider
	TestConfig      *config.TestConfig
}

func NewTestRuntime(t *testing.T) *TestRuntime {
	tc := config.NewTestConfig(t)
	claimStorage := claims.NewTestClaimProvider()
	credentialStorage := credentials.NewTestCredentialProvider(t, tc)
	parameterStorage := parameters.NewTestParameterProvider(t, tc)
	return NewTestRuntimeWithConfig(tc, claimStorage, credentialStorage, parameterStorage)
}

func NewTestRuntimeWithConfig(
	tc *config.TestConfig,
	testClaims claims.ClaimProvider,
	testCredentials credentials.TestCredentialProvider,
	testParameters parameters.TestParameterProvider) *TestRuntime {
	return &TestRuntime{
		TestConfig:      tc,
		TestCredentials: testCredentials,
		Runtime:         NewRuntime(tc.Config, testClaims, testCredentials, testParameters),
	}
}

func (t *TestRuntime) LoadBundle(bundleFile string) (*bundle.Bundle, error) {
	return t.Runtime.LoadBundle(bundleFile)
}

func (t *TestRuntime) Install(args ActionArguments) error {
	args.Driver = debugDriver
	return t.Runtime.Install(args)
}

func (t *TestRuntime) Upgrade(args ActionArguments) error {
	args.Driver = debugDriver
	return t.Runtime.Upgrade(args)
}

func (t *TestRuntime) Invoke(action string, args ActionArguments) error {
	args.Driver = debugDriver
	return t.Runtime.Invoke(action, args)
}

func (t *TestRuntime) Uninstall(args ActionArguments) error {
	args.Driver = debugDriver
	return t.Runtime.Uninstall(args)
}
