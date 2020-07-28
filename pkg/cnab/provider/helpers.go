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
	TestClaims      claims.TestClaimProvider
	TestCredentials credentials.TestCredentialProvider
	TestParameters  *parameters.TestParameterProvider
	TestConfig      *config.TestConfig
}

func NewTestRuntime(t *testing.T) *TestRuntime {
	tc := config.NewTestConfig(t)
	claimStorage := claims.NewTestClaimProvider(t)
	credentialStorage := credentials.NewTestCredentialProvider(t, tc)
	parameterStorage := parameters.NewTestParameterProvider(t, tc)
	return NewTestRuntimeWithConfig(tc, claimStorage, credentialStorage, parameterStorage)
}

func NewTestRuntimeWithConfig(tc *config.TestConfig, testClaims claims.TestClaimProvider, testCredentials credentials.TestCredentialProvider, testParameters parameters.TestParameterProvider) *TestRuntime {
	return &TestRuntime{
		TestConfig:      tc,
		TestClaims:      testClaims,
		TestCredentials: testCredentials,
		TestParameters:  &testParameters,
		Runtime:         NewRuntime(tc.Config, testClaims, testCredentials, testParameters),
	}
}

func (t *TestRuntime) LoadBundle(bundleFile string) (bundle.Bundle, error) {
	return t.Runtime.LoadBundle(bundleFile)
}

func (t *TestRuntime) Execute(args ActionArguments) error {
	args.Driver = debugDriver
	return t.Runtime.Execute(args)
}
