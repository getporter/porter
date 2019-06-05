package cache

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	kahn1dot0Hash  = "887e7e65e39277f8744bd00278760b06"
	kahn1dot01     = "deislabs/kubekahn:1.0"
	kahnlatestHash = "fd4bbe38665531d10bb653140842a370"
	kahnlatest     = "deislabs/kubekahn:latest"
)

func TestBundleTagHash(t *testing.T) {

	bid := getBundleID(kahn1dot01)
	require.Equal(t, kahn1dot0Hash, bid, "hashing the bundle ID twice should be the same")

	bid2 := getBundleID(kahnlatest)
	assert.NotEqual(t, kahn1dot0Hash, bid2, "different tags should result in different hashes")

}

func TestFindBundleCacheExists(t *testing.T) {
	cfg := config.NewTestConfig(t)
	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)
	c := New(cfg.Config)

	_, ok, err := c.FindBundle("deislabs/kubekahn:latest")
	assert.NoError(t, err, "the cache dir should exist, no error should have happened")
	assert.False(t, ok, "the bundle shouldn't exist")
}

func TestFindBundleCacheDoesNotExist(t *testing.T) {
	cfg := config.NewTestConfig(t)
	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)
	c := New(cfg.Config)

	_, ok, err := c.FindBundle("deislabs/kubekahn:latest")
	assert.NoError(t, err, "the cache dir doesn't exist, but this shouldn't be an error")
	assert.False(t, ok, "the bundle shouldn't exist")
}

func TestFindBundleBundleCached(t *testing.T) {
	cfg := config.NewTestConfig(t)
	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)
	expectedCacheDirectory := filepath.Join(cacheDir, kahn1dot0Hash)
	expectedCacheCNABDirectory := filepath.Join(expectedCacheDirectory, "cnab")
	expectedCacheFile := filepath.Join(expectedCacheCNABDirectory, "bundle.json")
	foundIt, err := cfg.Config.FileSystem.Exists(expectedCacheFile)
	require.True(t, foundIt, "test data not loaded")
	c := New(cfg.Config)
	path, ok, err := c.FindBundle(kahn1dot01)
	assert.NoError(t, err, "the cache dir should exist, no error should have happened")
	assert.True(t, ok, "the bundle should exist")
	assert.Equal(t, expectedCacheFile, path)
}

func TestFindBundleBundleNotCached(t *testing.T) {
	cfg := config.NewTestConfig(t)
	c := New(cfg.Config)
	path, ok, err := c.FindBundle(kahnlatest)
	assert.NoError(t, err, "the cache dir should exist, no error should have happened")
	assert.False(t, ok, "the bundle should not exist")
	assert.Empty(t, path, "should not have a path")
}

func TestCacheWriteNoCacheDir(t *testing.T) {
	cfg := config.NewTestConfig(t)
	cfg.TestContext.AddTestFile("testdata/cnab/bundle.json", "/cnab/bundle.json")
	b, err := cfg.FileSystem.ReadFile("/cnab/bundle.json")
	bun, err := bundle.ParseReader(bytes.NewBuffer(b))
	require.NoError(t, err, "bundle should have been valid")

	c := New(cfg.Config)
	path, err := c.StoreBundle(kahn1dot01, &bun)

	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	expectedCacheDirectory := filepath.Join(cacheDir, kahn1dot0Hash)
	expectedCacheCNABDirectory := filepath.Join(expectedCacheDirectory, "cnab")
	expectedCacheFile := filepath.Join(expectedCacheCNABDirectory, "bundle.json")

	assert.Equal(t, expectedCacheFile, path)
	assert.NoError(t, err, "storing bundle should have succeeded")
}

func TestCacheWriteCacheDirExists(t *testing.T) {
	cfg := config.NewTestConfig(t)
	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestFile("testdata/cnab/bundle.json", "/cnab/bundle.json")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)
	b, err := cfg.FileSystem.ReadFile("/cnab/bundle.json")
	bun, err := bundle.ParseReader(bytes.NewBuffer(b))
	require.NoError(t, err, "bundle should have been valid")

	c := New(cfg.Config)
	path, err := c.StoreBundle(kahn1dot01, &bun)

	expectedCacheDirectory := filepath.Join(cacheDir, kahn1dot0Hash)
	expectedCacheCNABDirectory := filepath.Join(expectedCacheDirectory, "cnab")
	expectedCacheFile := filepath.Join(expectedCacheCNABDirectory, "bundle.json")

	assert.Equal(t, expectedCacheFile, path)
	assert.NoError(t, err, "storing bundle should have succeeded")
}
