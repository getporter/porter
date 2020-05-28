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
	"get.porter.sh/porter/pkg/secrets"
	cnabcreds "github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type TestPorter struct {
	*Porter
	TestConfig      *config.TestConfig
	TestCredentials *credentials.TestCredentialProvider
	TestCache       *cache.TestCache

	// original directory where the test was being executed
	TestDir string

	// directory where the integration test is being executed
	BundleDir string

	// tempDirectories that need to be cleaned up at the end of the testRun
	cleanupDirs []string
}

// NewTestPorter initializes a porter test client, with the output buffered, and an in-memory file system.
func NewTestPorter(t *testing.T) *TestPorter {
	tc := config.NewTestConfig(t)
	testCredentials := credentials.NewTestCredentialProvider(t, tc)
	testParameters := parameters.NewTestParameterProvider(t, tc)
	testCache := cache.NewTestCache(cache.New(tc.Config))
	testClaims := claims.NewTestClaimProvider()

	p := New()
	p.Config = tc.Config
	p.Mixins = mixin.NewTestMixinProvider()
	p.Plugins = plugins.NewTestPluginProvider()
	p.Cache = testCache
	p.Builder = NewTestBuildProvider()
	p.Claims = testClaims
	p.Credentials = testCredentials
	p.CNAB = cnabprovider.NewTestRuntimeWithConfig(tc, testClaims, testCredentials, testParameters)

	return &TestPorter{
		Porter:          p,
		TestConfig:      tc,
		TestCredentials: &testCredentials,
		TestCache:       testCache,
	}
}

func (p *TestPorter) SetupIntegrationTest() {
	t := p.TestConfig.TestContext.T

	// Undo changes above to make a unit test friendly Porter, so we hit the host
	p.Porter = NewWithConfig(p.Config)
	p.NewCommand = exec.Command
	p.TestCredentials.SecretsStore = secrets.NewSecretStore(&host.SecretStore{})

	homeDir := p.UseFilesystem()
	p.TestConfig.SetupIntegrationTest(homeDir)
	bundleDir := p.CreateBundleDir()

	p.TestDir, _ = os.Getwd()
	err := os.Chdir(bundleDir)
	require.NoError(t, err)

	// Copy test credentials into porter home, with KUBECONFIG replaced properly
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

// UseFilesystem has porter's context use the OS filesystem instead of an in-memory filesystem
// Returns the temp porter home directory created for the test
func (p *TestPorter) UseFilesystem() string {
	p.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}

	homeDir, err := ioutil.TempDir("/tmp", "porter")
	require.NoError(p.T(), err)
	p.cleanupDirs = append(p.cleanupDirs, homeDir)

	return homeDir
}

func (p *TestPorter) CreateBundleDir() string {
	bundleDir, err := ioutil.TempDir("", "bundle")
	require.NoError(p.T(), err)

	p.BundleDir = bundleDir
	p.cleanupDirs = append(p.cleanupDirs, p.BundleDir)

	return bundleDir
}

func (p *TestPorter) T() *testing.T {
	return p.TestConfig.TestContext.T
}

func (p *TestPorter) CleanupIntegrationTest() {
	os.Unsetenv(config.EnvHOME)

	for _, dir := range p.cleanupDirs {
		p.FileSystem.RemoveAll(dir)
	}

	os.Chdir(p.TestDir)
}

type TestBuildProvider struct{}

func NewTestBuildProvider() *TestBuildProvider {
	return &TestBuildProvider{}
}
func (t *TestBuildProvider) BuildInvocationImage(manifest *manifest.Manifest) error {
	return nil
}
