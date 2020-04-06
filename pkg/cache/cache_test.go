package cache

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/docker/cnab-to-oci/relocation"
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

	_, _, ok, err := c.FindBundle("deislabs/kubekahn:latest")
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

	_, _, ok, err := c.FindBundle("deislabs/kubekahn:latest")
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
	path, _, ok, err := c.FindBundle(kahn1dot01)
	assert.NoError(t, err, "the cache dir should exist, no error should have happened")
	assert.True(t, ok, "the bundle should exist")
	assert.Equal(t, expectedCacheFile, path)
}

func TestFindBundleBundleNotCached(t *testing.T) {
	cfg := config.NewTestConfig(t)
	c := New(cfg.Config)
	path, _, ok, err := c.FindBundle(kahnlatest)
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
	var reloMap relocation.ImageRelocationMap
	path, _, err := c.StoreBundle(kahn1dot01, &bun, reloMap)

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
	var reloMap relocation.ImageRelocationMap
	path, _, err := c.StoreBundle(kahn1dot01, &bun, reloMap)

	expectedCacheDirectory := filepath.Join(cacheDir, kahn1dot0Hash)
	expectedCacheCNABDirectory := filepath.Join(expectedCacheDirectory, "cnab")
	expectedCacheFile := filepath.Join(expectedCacheCNABDirectory, "bundle.json")

	assert.Equal(t, expectedCacheFile, path)
	assert.NoError(t, err, "storing bundle should have succeeded")
}

func TestStoreRelocationMapping(t *testing.T) {

	cfg := config.NewTestConfig(t)
	home, _ := cfg.Config.GetHomeDir()
	cacheDir := filepath.Join(home, "cache")

	tests := []struct {
		name              string
		relocationMapping relocation.ImageRelocationMap
		tag               string
		bundle            *bundle.Bundle
		wantedReloPath    string
		err               error
	}{
		{
			name:   "relocation file gets a path",
			bundle: &bundle.Bundle{},
			tag:    kahn1dot01,
			relocationMapping: relocation.ImageRelocationMap{
				"asd": "asdf",
			},
			wantedReloPath: filepath.Join(cacheDir, kahn1dot0Hash, "cnab", "relocation-mapping.json"),
		},
		{
			name:           "no relocation file gets no path",
			tag:            kahnlatest,
			bundle:         &bundle.Bundle{},
			wantedReloPath: "",
		},
	}

	c := New(cfg.Config)
	for _, test := range tests {
		_, reloPath, err := c.StoreBundle(test.tag, test.bundle, test.relocationMapping)
		assert.NoError(t, err, fmt.Sprintf("didn't expect storage error for test %s", test.name))
		assert.Equal(t, test.wantedReloPath, reloPath, "didn't get expected path for store on test: %s", test.name)
		_, fetchedPath, _, err := c.FindBundle(test.tag)
		assert.Equal(t, test.wantedReloPath, fetchedPath, "didn't get expected path for load on test: %s", test.name)

	}
}

func TestStoreManifest(t *testing.T) {

	cfg := config.NewTestConfig(t)
	home, _ := cfg.Config.GetHomeDir()
	cacheDir := filepath.Join(home, "cache")
	cachedManifestPath := filepath.Join(cacheDir, kahn1dot0Hash, "porter.yaml")

	tests := []struct {
		name                string
		tag                 string
		bundle              *bundle.Bundle
		shouldCacheManifest bool
		err                 error
	}{
		{
			name: "embedded manifest",
			bundle: &bundle.Bundle{
				Custom: map[string]interface{}{
					"sh.porter": map[string]interface{}{
						"manifest": "abc123",
					},
				},
			},
			tag:                 kahn1dot01,
			shouldCacheManifest: true,
		},
		{
			name:   "no embedded manifest",
			tag:    kahnlatest,
			bundle: &bundle.Bundle{},
		},
	}

	c := New(cfg.Config)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := c.StoreBundle(test.tag, test.bundle, nil)
			require.NoError(t, err, "StoreBundle failed")

			cachedManifestExists, _ := cfg.FileSystem.Exists(cachedManifestPath)
			if test.shouldCacheManifest {
				assert.True(t, cachedManifestExists, "Expected the porter.yaml manifest to be cached but it wasn't")
			} else {
				assert.False(t, cachedManifestExists, "Expected no porter.yaml manifest to be cached but one was cached anyway. Not sure what happened there...")
			}
		})
	}
}
