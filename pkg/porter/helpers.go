package porter

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	instancestorage "get.porter.sh/porter/pkg/instance-storage"

	"get.porter.sh/porter/pkg/manifest"

	buildprovider "get.porter.sh/porter/pkg/build/provider"
	"get.porter.sh/porter/pkg/cache"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/mixin"
	mixinprovider "get.porter.sh/porter/pkg/mixin/provider"
	"github.com/deislabs/cnab-go/bundle"
	"github.com/docker/cnab-to-oci/relocation"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

type TestPorter struct {
	*Porter
	TestConfig *config.TestConfig

	// original directory where the test was being executed
	TestDir string

	// tempDirectories that need to be cleaned up at the end of the testRun
	cleanupDirs []string
}

// NewTestPorter initializes a porter test client, with the output buffered, and an in-memory file system.
func NewTestPorter(t *testing.T) *TestPorter {
	tc := config.NewTestConfig(t)
	p := New()
	p.Config = tc.Config
	p.Mixins = &mixin.TestMixinProvider{}
	p.Cache = cache.New(tc.Config)
	p.Builder = NewTestBuildProvider()
	p.InstanceStorage = instancestorage.NewTestInstanceStorageProvider()
	p.CNAB = cnabprovider.NewRuntime(tc.Config, p.InstanceStorage)

	return &TestPorter{
		Porter:     p,
		TestConfig: tc,
	}
}

func (p *TestPorter) SetupIntegrationTest() {
	t := p.TestConfig.TestContext.T

	p.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}
	p.NewCommand = exec.Command
	p.Builder = buildprovider.NewDockerBuilder(p.Context)
	p.Mixins = mixinprovider.NewFileSystem(p.Config)

	// Set up porter and the bundle inside of a temp directory
	homeDir, err := ioutil.TempDir("/tmp", "porter")
	require.NoError(t, err)
	p.cleanupDirs = append(p.cleanupDirs, homeDir)
	p.TestConfig.SetupIntegrationTest(homeDir)

	bundleDir, err := ioutil.TempDir("", "bundle")
	require.NoError(t, err)
	p.cleanupDirs = append(p.cleanupDirs, homeDir)

	p.TestDir, _ = os.Getwd()
	err = os.Chdir(bundleDir)
	require.NoError(t, err)

	// Copy test credentials into porter home
	credsDir, _ := p.GetCredentialsDir()
	p.FileSystem.Mkdir(credsDir, 0755)
	ciCredsPath := filepath.Join(credsDir, "ci.yaml")
	err = p.CopyFile(filepath.Join(p.TestDir, "../build/testdata/credentials/ci.yaml"), ciCredsPath)
	require.NoError(t, err, "could not copy credentials file")

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home := os.Getenv("HOME")
		kubeconfig = filepath.Join(home, ".kube/config")
	}

	ciCredsB, _ := p.FileSystem.ReadFile(ciCredsPath)
	ciCredsB = []byte(strings.Replace(string(ciCredsB), "KUBECONFIGPATH", kubeconfig, -1))
	err = p.FileSystem.WriteFile(ciCredsPath, ciCredsB, 0755)
	require.NoError(t, err, "could not update the credentials file with KUBECONFIG")
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
