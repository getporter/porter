package porter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/claims"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
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
)

type TestPorter struct {
	*Porter
	TestConfig      *config.TestConfig
	TestClaims      claims.TestClaimProvider
	TestCredentials *credentials.TestCredentialProvider
	TestParameters  *parameters.TestParameterProvider
	TestCache       *cache.TestCache
	TestRegistry    *cnabtooci.TestRegistry

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
	testRegistry := cnabtooci.NewTestRegistry()

	p := New()
	p.Config = tc.Config
	p.Mixins = mixin.NewTestMixinProvider(tc)
	p.Plugins = plugins.NewTestPluginProvider()
	p.Cache = testCache
	p.Builder = NewTestBuildProvider()
	p.Claims = testClaims
	p.Credentials = testCredentials
	p.Parameters = testParameters
	p.CNAB = cnabprovider.NewTestRuntimeWithConfig(tc, testClaims, testCredentials, testParameters)
	p.Registry = testRegistry

	return &TestPorter{
		Porter:          p,
		TestConfig:      tc,
		TestClaims:      testClaims,
		TestCredentials: &testCredentials,
		TestParameters:  &testParameters,
		TestCache:       testCache,
		TestRegistry:    testRegistry,
	}
}

func (p *TestPorter) SetupIntegrationTest() {
	t := p.TestConfig.TestContext.T

	// Undo changes above to make a unit test friendly Porter, so we hit the host
	p.Porter = NewWithConfig(p.Config)

	/*
		// Update test providers to use the instances we just reset above
		// We mostly don't use test providers for integration tests, but a few provide
		// useful helper methods that are still nice to have.
		hostSecrets := &host.SecretStore{}
		p.TestCredentials.SecretsStore = secrets.NewSecretStore(hostSecrets)
		p.TestParameters.SecretsStore = secrets.NewSecretStore(hostSecrets)
	*/

	// Run the test in a temp directory
	testDir, homeDir := p.TestConfig.SetupIntegrationTest()
	p.TestDir = testDir
	p.CreateBundleDir()

	// Copy test credentials into porter home, with KUBECONFIG replaced properly
	p.AddTestFile("../build/testdata/schema.json", filepath.Join(homeDir, "schema.json"))
	kubeconfig := filepath.ToSlash(p.Getenv("KUBECONFIG"))
	if kubeconfig == "" {
		home := p.Getenv("HOME")
		kubeconfig = path.Join(home, ".kube/config")
	}
	ciCredsPath := filepath.Join(p.TestDir, "../build/testdata/credentials/ci.json")
	ciCredsB, err := p.FileSystem.ReadFile(ciCredsPath)

	require.NoError(t, err, "could not read test credentials %s", ciCredsPath)
	// update the kubeconfig reference in the credentials to match what's on people's dev machine
	ciCredsB = []byte(strings.Replace(string(ciCredsB), "KUBECONFIGPATH", kubeconfig, -1))
	var testCreds cnabcreds.CredentialSet
	err = json.Unmarshal(ciCredsB, &testCreds)
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

func (p *TestPorter) AddTestDirectory(src string, dest string) {
	if !filepath.IsAbs(src) {
		src = filepath.Join(p.TestDir, src)
	}

	p.TestConfig.TestContext.AddTestDirectory(src, dest)
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
	p.Chdir(bundleDir)
	p.TestConfig.TestContext.AddCleanupDir(p.BundleDir)

	return bundleDir
}

func (p *TestPorter) T() *testing.T {
	return p.TestConfig.TestContext.T
}

func (p *TestPorter) CleanupIntegrationTest() {
	p.TestConfig.TestContext.Cleanup()
}

func (p *TestPorter) ReadBundle(path string) bundle.Bundle {
	bunD, err := ioutil.ReadFile(path)
	require.NoError(p.T(), err, "ReadFile failed for %s", path)

	bun, err := bundle.Unmarshal(bunD)
	require.NoError(p.T(), err, "Unmarshal failed for bundle at %s", path)

	return *bun
}

func (p *TestPorter) RandomString(len int) string {
	rand.Seed(time.Now().UnixNano())
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		//A=97 and Z = 97+25
		bytes[i] = byte(97 + rand.Intn(25))
	}
	return string(bytes)
}

// AddTestBundleDir into the test bundle directory and give it a unique name
// to avoid collisions with other tests running in parallel.
func (p *TestPorter) AddTestBundleDir(bundleDir string, generateUniqueName bool) string {
	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.TestDir, bundleDir), p.BundleDir)

	testManifest := filepath.Join(p.BundleDir, config.Name)
	m, err := manifest.LoadManifestFrom(p.Context, testManifest)
	require.NoError(p.T(), err)

	if !generateUniqueName {
		return m.Name
	}

	e := manifest.NewEditor(p.Context)
	err = e.ReadFile(testManifest)
	require.NoError(p.T(), err)

	uniqueName := fmt.Sprintf("%s-%s", m.Name, p.RandomString(5))
	err = e.SetValue("name", uniqueName)
	require.NoError(p.T(), err)

	err = e.WriteFile(testManifest)
	require.NoError(p.T(), err)

	return uniqueName
}

// Copy a bundle from the Porter repository into the test's cache
func (p *TestPorter) CacheTestBundle(srcDir string, tag string) {
	cacheDir := p.Cache.GetCacheDir()
	cb := cache.CachedBundle{Tag: tag}
	destDir := filepath.Join(cacheDir, cb.GetBundleID())
	p.AddTestDirectory(srcDir, destDir)
	p.FileSystem.Rename(filepath.Join(destDir, ".cnab"), filepath.Join(destDir, "cnab"))
}

type TestBuildProvider struct {
}

func NewTestBuildProvider() *TestBuildProvider {
	return &TestBuildProvider{}
}

func (t *TestBuildProvider) BuildInvocationImage(manifest *manifest.Manifest) error {
	return nil
}
