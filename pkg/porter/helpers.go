package porter

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/require"
)

type TestPorter struct {
	*Porter
	TestConfig      *config.TestConfig
	TestStore       storage.TestStore
	TestClaims      *claims.TestClaimProvider
	TestCredentials *credentials.TestCredentialProvider
	TestParameters  *parameters.TestParameterProvider
	TestCache       *cache.TestCache
	TestRegistry    *cnabtooci.TestRegistry

	// original directory where the test was being executed
	TestDir string

	// directory where the integration test is being executed
	BundleDir string

	// root of the repository
	// Helps us avoid hard coding relative paths from test directories, which easily break when tests are moved
	RepoRoot string

	// The root test context created by NewTestPorter
	RootContext context.Context

	// The root log span created by NewTestPorter
	RootSpan tracing.TraceLogger
}

// NewTestPorter initializes a porter test client, with the output buffered, and an in-memory file system.
func NewTestPorter(t *testing.T) *TestPorter {
	tc := config.NewTestConfig(t)
	testStore := storage.NewTestStore(tc.TestContext)
	testCredentials := credentials.NewTestCredentialProviderFor(t, testStore)
	testParameters := parameters.NewTestParameterProviderFor(t, testStore)
	testCache := cache.NewTestCache(cache.New(tc.Config))
	testClaims := claims.NewTestClaimProviderFor(t, testStore)
	testRegistry := cnabtooci.NewTestRegistry()

	p := NewFor(tc.Config, testStore)
	p.Config = tc.Config
	p.Mixins = mixin.NewTestMixinProvider()
	p.Plugins = plugins.NewTestPluginProvider()
	p.Cache = testCache
	p.builder = NewTestBuildProvider()
	p.Claims = testClaims
	p.Credentials = testCredentials
	p.Parameters = testParameters
	p.CNAB = cnabprovider.NewTestRuntimeFor(tc, testClaims, testCredentials, testParameters)
	p.Registry = testRegistry

	tp := TestPorter{
		Porter:          p,
		TestConfig:      tc,
		TestStore:       testStore,
		TestClaims:      testClaims,
		TestCredentials: testCredentials,
		TestParameters:  testParameters,
		TestCache:       testCache,
		TestRegistry:    testRegistry,
		RepoRoot:        tc.TestContext.FindRepoRoot(),
	}

	// Start a tracing span for the test, so that we can capture logs
	tp.RootContext, tp.RootSpan = p.StartRootSpan(context.Background(), t.Name())

	return &tp
}

func (p *TestPorter) Teardown() error {
	err := p.TestStore.Teardown()
	p.TestConfig.TestContext.Teardown()
	p.RootSpan.EndSpan()
	return err
}

func (p *TestPorter) SetupIntegrationTest() {
	t := p.TestConfig.TestContext.T

	// Undo changes above to make a unit test friendly Porter, so we hit the host
	p.Porter = NewFor(p.Config, p.TestStore)

	// Run the test in a temp directory
	testDir, _ := p.TestConfig.SetupIntegrationTest()
	p.TestDir = testDir
	p.CreateBundleDir()

	// Write out a storage schema so that we don't trigger a migration check
	err := p.Storage.WriteSchema()
	require.NoError(t, err, "failed to set the storage schema")

	// Load test credentials, with KUBECONFIG replaced properly
	kubeconfig := filepath.Join(p.RepoRoot, "kind.config")
	ciCredsPath := filepath.Join(p.RepoRoot, "build/testdata/credentials/ci.json")
	ciCredsB, err := p.FileSystem.ReadFile(ciCredsPath)
	require.NoError(t, err, "could not read test credentials %s", ciCredsPath)
	// update the kubeconfig reference in the credentials to match what's on people's dev machine
	ciCredsB = []byte(strings.Replace(string(ciCredsB), "KUBECONFIGPATH", kubeconfig, -1))
	var testCreds credentials.CredentialSet
	err = encoding.UnmarshalYaml(ciCredsB, &testCreds)
	require.NoError(t, err, "could not unmarshal test credentials %s", ciCredsPath)
	err = p.Credentials.UpsertCredentialSet(testCreds)
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
	p.Chdir(bundleDir)
	p.TestConfig.TestContext.AddCleanupDir(p.BundleDir)

	return bundleDir
}

func (p *TestPorter) T() *testing.T {
	return p.TestConfig.TestContext.T
}

func (p *TestPorter) ReadBundle(path string) cnab.ExtendedBundle {
	bunD, err := ioutil.ReadFile(path)
	require.NoError(p.T(), err, "ReadFile failed for %s", path)

	bun, err := bundle.Unmarshal(bunD)
	require.NoError(p.T(), err, "Unmarshal failed for bundle at %s", path)

	return cnab.ExtendedBundle{*bun}
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
	if !filepath.IsAbs(bundleDir) {
		bundleDir = filepath.Join(p.TestDir, bundleDir)
	}
	p.TestConfig.TestContext.AddTestDirectory(bundleDir, p.BundleDir)

	testManifest := filepath.Join(p.BundleDir, config.Name)
	m, err := manifest.LoadManifestFrom(p.RootContext, p.Config, testManifest)
	require.NoError(p.T(), err)

	if !generateUniqueName {
		return m.Name
	}

	e := yaml.NewEditor(p.Context)
	err = e.ReadFile(testManifest)
	require.NoError(p.T(), err)

	uniqueName := fmt.Sprintf("%s-%s", m.Name, p.RandomString(5))
	err = e.SetValue("name", uniqueName)
	require.NoError(p.T(), err)

	err = e.WriteFile(testManifest)
	require.NoError(p.T(), err)

	return uniqueName
}

// CompareGoldenFile checks if the specified string matches the content of a golden test file.
// When they are different and PORTER_UPDATE_TEST_FILES is true, the file is updated to match
// the new test output.
func (p *TestPorter) CompareGoldenFile(goldenFile string, got string) {
	p.TestConfig.TestContext.CompareGoldenFile(goldenFile, got)
}

type TestBuildProvider struct {
}

func NewTestBuildProvider() *TestBuildProvider {
	return &TestBuildProvider{}
}

func (t *TestBuildProvider) BuildInvocationImage(ctx context.Context, manifest *manifest.Manifest, opts build.BuildImageOptions) error {
	return nil
}

func (t *TestBuildProvider) TagInvocationImage(ctx context.Context, origTag, newTag string) error {
	return nil
}
