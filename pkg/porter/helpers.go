package porter

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	buildprovider "get.porter.sh/porter/pkg/build/provider"
	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/claims"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	mixinprovider "get.porter.sh/porter/pkg/mixin/provider"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/bundle"
	cnabcreds "github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/docker/cnab-to-oci/relocation"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type TestPorter struct {
	*Porter
	TestConfig      *config.TestConfig
	TestCredentials *credentials.TestCredentialProvider

	// original directory where the test was being executed
	TestDir string

	// tempDirectories that need to be cleaned up at the end of the testRun
	cleanupDirs []string
}

// NewTestPorter initializes a porter test client, with the output buffered, and an in-memory file system.
func NewTestPorter(t *testing.T) *TestPorter {
	tc := config.NewTestConfig(t)
	testCredentials := credentials.NewTestCredentialProvider(t, tc)
	p := New()
	p.Config = tc.Config
	p.Mixins = &mixin.TestMixinProvider{}
	p.Plugins = &plugins.TestPluginProvider{}
	p.Cache = cache.New(tc.Config)
	p.Builder = NewTestBuildProvider()
	p.Claims = claims.NewTestClaimProvider()
	p.Credentials = testCredentials
	p.CNAB = cnabprovider.NewRuntime(tc.Config, p.Claims, p.Credentials)

	return &TestPorter{
		Porter:          p,
		TestConfig:      tc,
		TestCredentials: &testCredentials,
	}
}

func (p *TestPorter) SetupIntegrationTest() {
	t := p.TestConfig.TestContext.T

	p.NewCommand = exec.Command
	p.Builder = buildprovider.NewDockerBuilder(p.Context)
	p.Mixins = mixinprovider.NewFileSystem(p.Config)
	p.TestCredentials.SecretsStore = secrets.NewSecretStore(&host.SecretStore{})

	homeDir := p.UseFilesystem()
	p.TestConfig.SetupIntegrationTest(homeDir)

	bundleDir, err := ioutil.TempDir("", "bundle")
	require.NoError(t, err)
	p.cleanupDirs = append(p.cleanupDirs, homeDir)

	p.TestDir, _ = os.Getwd()
	err = os.Chdir(bundleDir)
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

// If you seek a mock cache for testing, use this
type mockCache struct {
	findBundleMock        func(string) (string, string, bool, error)
	storeBundleMock       func(string, *bundle.Bundle, relocation.ImageRelocationMap) (string, string, error)
	getBundleCacheDirMock func() (string, error)
}

func (b *mockCache) FindBundle(tag string) (string, string, bool, error) {
	return b.findBundleMock(tag)
}

func (b *mockCache) StoreBundle(tag string, bun *bundle.Bundle, relo relocation.ImageRelocationMap) (string, string, error) {
	return b.storeBundleMock(tag, bun, relo)
}

func (b *mockCache) GetCacheDir() (string, error) {
	return b.GetCacheDir()
}

type TestCNABProvider struct {
}

func NewTestCNABProvider() *TestCNABProvider {
	return &TestCNABProvider{}
}

func (t *TestCNABProvider) LoadBundle(bundleFile string, insecure bool) (*bundle.Bundle, error) {
	b := &bundle.Bundle{
		Name: "testbundle",
		Credentials: map[string]bundle.Credential{
			"name": {
				Location: bundle.Location{
					EnvironmentVariable: "BLAH",
				},
			},
		},
	}
	return b, nil
}

func (t *TestCNABProvider) Install(arguments cnabprovider.ActionArguments) error {
	return nil
}

func (t *TestCNABProvider) Upgrade(arguments cnabprovider.ActionArguments) error {
	return nil
}

func (t *TestCNABProvider) Invoke(action string, arguments cnabprovider.ActionArguments) error {
	return nil
}

func (t *TestCNABProvider) Uninstall(arguments cnabprovider.ActionArguments) error {
	return nil
}

type TestBuildProvider struct{}

func NewTestBuildProvider() *TestBuildProvider {
	return &TestBuildProvider{}
}
func (t *TestBuildProvider) BuildInvocationImage(manifest *manifest.Manifest) error {
	return nil
}
