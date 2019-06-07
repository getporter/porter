package porter

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/cache"
	"github.com/deislabs/porter/pkg/config"
	execmixin "github.com/deislabs/porter/pkg/exec"
	"github.com/deislabs/porter/pkg/mixin"
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
	p.Mixins = &TestMixinProvider{}
	p.Cache = cache.New(tc.Config)
	return &TestPorter{
		Porter:     p,
		TestConfig: tc,
	}
}

func (p *TestPorter) SetupIntegrationTest() {
	t := p.TestConfig.TestContext.T

	p.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}
	p.NewCommand = exec.Command

	// Set up porter and the bundle inside of a temp directory
	homeDir, err := ioutil.TempDir("", "porter")
	require.NoError(t, err)
	p.cleanupDirs = append(p.cleanupDirs, homeDir)
	p.TestConfig.SetupIntegrationTest(homeDir)

	bundleDir, err := ioutil.TempDir("", "bundle")
	require.NoError(t, err)
	p.cleanupDirs = append(p.cleanupDirs, homeDir)

	p.TestDir, _ = os.Getwd()
	err = os.Chdir(bundleDir)
	require.NoError(t, err)
}

func (p *TestPorter) CleanupIntegrationTest() {
	os.Unsetenv(config.EnvHOME)

	for _, dir := range p.cleanupDirs {
		p.FileSystem.RemoveAll(dir)
	}

	os.Chdir(p.TestDir)
}

// TODO: use this later to not actually execute a mixin during a unit test
type TestMixinProvider struct {
}

func (p *TestMixinProvider) List() ([]mixin.Metadata, error) {
	mixins := []mixin.Metadata{
		{Name: "exec"},
	}
	return mixins, nil
}

func (p *TestMixinProvider) GetSchema(m mixin.Metadata) (string, error) {
	t := execmixin.NewSchemaBox()
	return t.FindString("exec.json")
}

func (p *TestMixinProvider) GetVersion(m mixin.Metadata) (string, error) {
	return "exec mixin v1.0 (abc123)", nil
}

func (p *TestMixinProvider) Install(o mixin.InstallOptions) (mixin.Metadata, error) {
	return mixin.Metadata{Name: "exec", Dir: "~/.porter/mixins/exec"}, nil
}

// If you seek a mock cache for testing, use this
type mockCache struct {
	findBundleMock        func(string) (string, bool, error)
	storeBundleMock       func(string, *bundle.Bundle) (string, error)
	getBundleCacheDirMock func() (string, error)
}

func (b *mockCache) FindBundle(tag string) (string, bool, error) {
	return b.findBundleMock(tag)
}

func (b *mockCache) StoreBundle(tag string, bun *bundle.Bundle) (string, error) {
	return b.storeBundleMock(tag, bun)
}

func (b *mockCache) GetCacheDir() (string, error) {
	return b.GetCacheDir()
}
