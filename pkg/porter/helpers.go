package porter

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/claims"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/plugins"
	"github.com/cnabio/cnab-go/bundle"
	cnabcreds "github.com/cnabio/cnab-go/credentials"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type TestPorter struct {
	*Porter
	TestConfig      *config.TestConfig
	TestClaims      claims.TestClaimProvider
	TestCredentials *credentials.TestCredentialProvider
	TestParameters  *parameters.TestParameterProvider
	TestCache       *cache.TestCache

	// original directory where the test was being executed
	TestDir string

	// directory where the integration test is being executed
	BundleDir string
}

// NewTestPorter initializes a porter test client, with the output buffered, and an in-memory file system.
func NewTestPorter(t *testing.T) *TestPorter {
	tc := config.NewTestConfig(t)
	testCredentials := credentials.NewTestCredentialProvider(t, tc)
	testParameters := parameters.NewTestParameterProvider(t, tc)
	testCache := cache.NewTestCache(cache.New(tc.Config))
	testClaims := claims.NewTestClaimProvider(t)

	p := New()
	p.Config = tc.Config
	p.Mixins = mixin.NewTestMixinProvider()
	p.Plugins = plugins.NewTestPluginProvider()
	p.Cache = testCache
	p.Builder = NewTestBuildProvider()
	p.Claims = testClaims
	p.Credentials = testCredentials
	p.Parameters = testParameters
	p.CNAB = cnabprovider.NewTestRuntimeWithConfig(tc, testClaims, testCredentials, testParameters)

	return &TestPorter{
		Porter:          p,
		TestConfig:      tc,
		TestClaims:      testClaims,
		TestCredentials: &testCredentials,
		TestParameters:  &testParameters,
		TestCache:       testCache,
	}
}

func (p *TestPorter) SetupIntegrationTest() {
	t := p.TestConfig.TestContext.T

	// Undo changes above to make a unit test friendly Porter, so we hit the host
	p.Porter = NewWithConfig(p.Config)
	p.NewCommand = exec.Command

	/*
		// Update test providers to use the instances we just reset above
		// We mostly don't use test providers for integration tests, but a few provide
		// useful helper methods that are still nice to have.
		hostSecrets := &host.SecretStore{}
		p.TestCredentials.SecretsStore = secrets.NewSecretStore(hostSecrets)
		p.TestParameters.SecretsStore = secrets.NewSecretStore(hostSecrets)
	*/

	// Run the test in a temp directory
	homeDir := p.TestConfig.TestContext.UseFilesystem()
	p.TestConfig.SetupIntegrationTest(homeDir)
	bundleDir := p.CreateBundleDir()

	p.TestDir, _ = os.Getwd()
	err := os.Chdir(bundleDir)
	require.NoError(t, err)

	// Copy test credentials into porter home, with KUBECONFIG replaced properly
	p.AddTestFile("../build/testdata/schema.json", filepath.Join(homeDir, "schema.json"))
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home := os.Getenv("HOME")
		kubeconfig = filepath.Join(home, ".kube/config")
	}
	ciCredsPath := filepath.Join(p.TestDir, "../build/testdata/credentials/ci.json")
	ciCredsB, err := p.FileSystem.ReadFile(ciCredsPath)
	require.NoError(t, err, "could not read test credentials %s", ciCredsPath)
	// update the kubeconfig reference in the credentials to match what's on people's dev machine
	ciCredsB = []byte(strings.Replace(string(ciCredsB), "KUBECONFIGPATH", kubeconfig, -1))
	var testCreds cnabcreds.CredentialSet
	err = yaml.Unmarshal(ciCredsB, &testCreds)
	require.NoError(t, err, "could not unmarshal test credentials %s", ciCredsPath)
	err = p.Credentials.Save(testCreds)
	require.NoError(t, err, "could not save test credentials")
}

func (p *TestPorter) AddTestFile(src string, dest string) {
	if !filepath.IsAbs(src) {
		src = filepath.Join(p.TestDir, src)
	}

	p.TestConfig.TestContext.AddTestFile(src, dest)
}

type TestDriver struct {
	Name     string
	Filepath string
}

func (p *TestPorter) AddTestDriver(driver TestDriver) string {
	if !filepath.IsAbs(driver.Filepath) {
		driver.Filepath = filepath.Join(p.TestDir, driver.Filepath)
	}

	return p.TestConfig.TestContext.AddTestDriver(driver.Filepath, driver.Name)
}

func (p *TestPorter) CreateBundleDir() string {
	bundleDir, err := ioutil.TempDir("", "bundle")
	require.NoError(p.T(), err)

	p.BundleDir = bundleDir
	p.TestConfig.TestContext.AddCleanupDir(p.BundleDir)

	return bundleDir
}

func (p *TestPorter) T() *testing.T {
	return p.TestConfig.TestContext.T
}

func (p *TestPorter) CleanupIntegrationTest() {
	os.Unsetenv(config.EnvHOME)

	p.TestConfig.TestContext.Cleanup()

	os.Chdir(p.TestDir)
}

func (p *TestPorter) ReadBundle(path string) bundle.Bundle {
	bunD, err := ioutil.ReadFile(path)
	require.NoError(p.T(), err, "ReadFile failed for %s", path)

	bun, err := bundle.Unmarshal(bunD)
	require.NoError(p.T(), err, "Unmarshal failed for bundle at %s", path)

	return *bun
}

type TestBuildProvider struct{}

func NewTestBuildProvider() *TestBuildProvider {
	return &TestBuildProvider{}
}

func (t *TestBuildProvider) BuildInvocationImage(manifest *manifest.Manifest) error {
	return nil
}
