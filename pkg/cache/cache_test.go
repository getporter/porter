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
	cb, ok, err := c.FindBundle(kahn1dot01)
	assert.NoError(t, err, "the cache dir should exist, no error should have happened")
	assert.True(t, ok, "the bundle should exist")
	assert.Equal(t, expectedCacheFile, cb.BundlePath)
}

func TestFindBundleBundleNotCached(t *testing.T) {
	cfg := config.NewTestConfig(t)
	c := New(cfg.Config)
	cb, ok, err := c.FindBundle(kahnlatest)
	assert.NoError(t, err, "the cache dir should exist, no error should have happened")
	assert.False(t, ok, "the bundle should not exist")
	assert.Empty(t, cb.BundlePath, "should not have a path")
}

func TestCacheWriteNoCacheDir(t *testing.T) {
	cfg := config.NewTestConfig(t)
	cfg.TestContext.AddTestFile("testdata/cnab/bundle.json", "/cnab/bundle.json")
	b, err := cfg.FileSystem.ReadFile("/cnab/bundle.json")
	bun, err := bundle.ParseReader(bytes.NewBuffer(b))
	require.NoError(t, err, "bundle should have been valid")

	c := New(cfg.Config)
	cb, err := c.StoreBundle(kahn1dot01, bun, nil)

	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	expectedCacheDirectory := filepath.Join(cacheDir, kahn1dot0Hash)
	expectedCacheCNABDirectory := filepath.Join(expectedCacheDirectory, "cnab")
	expectedCacheFile := filepath.Join(expectedCacheCNABDirectory, "bundle.json")

	assert.Equal(t, expectedCacheFile, cb.BundlePath)
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
	cb, err := c.StoreBundle(kahn1dot01, bun, &reloMap)

	expectedCacheDirectory := filepath.Join(cacheDir, kahn1dot0Hash)
	expectedCacheCNABDirectory := filepath.Join(expectedCacheDirectory, "cnab")
	expectedCacheFile := filepath.Join(expectedCacheCNABDirectory, "bundle.json")

	assert.Equal(t, expectedCacheFile, cb.BundlePath)
	assert.NoError(t, err, "storing bundle should have succeeded")
}

func TestStoreRelocationMapping(t *testing.T) {

	cfg := config.NewTestConfig(t)
	home, _ := cfg.Config.GetHomeDir()
	cacheDir := filepath.Join(home, "cache")

	tests := []struct {
		name              string
		relocationMapping *relocation.ImageRelocationMap
		tag               string
		bundle            bundle.Bundle
		wantedReloPath    string
		err               error
	}{
		{
			name:   "relocation file gets a path",
			bundle: bundle.Bundle{},
			tag:    kahn1dot01,
			relocationMapping: &relocation.ImageRelocationMap{
				"asd": "asdf",
			},
			wantedReloPath: filepath.Join(cacheDir, kahn1dot0Hash, "cnab", "relocation-mapping.json"),
		},
		{
			name:           "no relocation file gets no path",
			tag:            kahnlatest,
			bundle:         bundle.Bundle{},
			wantedReloPath: "",
		},
	}

	c := New(cfg.Config)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cb, err := c.StoreBundle(test.tag, test.bundle, test.relocationMapping)
			assert.NoError(t, err, fmt.Sprintf("didn't expect storage error for test %s", test.name))
			assert.Equal(t, test.wantedReloPath, cb.RelocationFilePath, "didn't get expected path for store")

			cb, _, err = c.FindBundle(test.tag)
			assert.Equal(t, test.wantedReloPath, cb.RelocationFilePath, "didn't get expected path for load")
		})
	}
}

func TestStoreManifest(t *testing.T) {

	cfg := config.NewTestConfig(t)
	home, _ := cfg.Config.GetHomeDir()
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)

	tests := []struct {
		name                string
		tag                 string
		bundle              bundle.Bundle
		shouldCacheManifest bool
		err                 error
	}{
		{
			name: "embedded manifest",
			bundle: bundle.Bundle{
				Custom: map[string]interface{}{
					"sh.porter": map[string]interface{}{
						"manifest": "bmFtZTogSEVMTE9fQ1VTVE9NCnZlcnNpb246IDAuMS4wCmRlc2NyaXB0aW9uOiAiQSBidW5kbGUgd2l0aCBhIGN1c3RvbSBhY3Rpb24iCnRhZzogZ2V0cG9ydGVyL3BvcnRlci1oZWxsbzp2MC4xLjAKaW52b2NhdGlvbkltYWdlOiBnZXRwb3J0ZXIvcG9ydGVyLWhlbGxvLWluc3RhbGxlcjowLjEuMAoKY3JlZGVudGlhbHM6CiAgLSBuYW1lOiBteS1maXJzdC1jcmVkCiAgICBlbnY6IE1ZX0ZJUlNUX0NSRUQKICAtIG5hbWU6IG15LXNlY29uZC1jcmVkCiAgICBkZXNjcmlwdGlvbjogIk15IHNlY29uZCBjcmVkIgogICAgcGF0aDogL3BhdGgvdG8vbXktc2Vjb25kLWNyZWQKCmltYWdlczogCiAgIHNvbWV0aGluZzoKICAgICAgZGVzY3JpcHRpb246ICJhbiBpbWFnZSIKICAgICAgaW1hZ2VUeXBlOiAiZG9ja2VyIgogICAgICByZXBvc2l0b3J5OiAiZ2V0cG9ydGVyL2JvbyIKCnBhcmFtZXRlcnM6CiAgLSBuYW1lOiBteS1maXJzdC1wYXJhbQogICAgdHlwZTogaW50ZWdlcgogICAgZGVmYXVsdDogOQogICAgZW52OiBNWV9GSVJTVF9QQVJBTQogICAgYXBwbHlUbzoKICAgICAgLSAiaW5zdGFsbCIKICAtIG5hbWU6IG15LXNlY29uZC1wYXJhbQogICAgZGVzY3JpcHRpb246ICJNeSBzZWNvbmQgcGFyYW1ldGVyIgogICAgdHlwZTogc3RyaW5nCiAgICBkZWZhdWx0OiBzcHJpbmctbXVzaWMtZGVtbwogICAgcGF0aDogL3BhdGgvdG8vbXktc2Vjb25kLXBhcmFtCiAgICBzZW5zaXRpdmU6IHRydWUKCm91dHB1dHM6CiAgLSBuYW1lOiBteS1maXJzdC1vdXRwdXQKICAgIHR5cGU6IHN0cmluZwogICAgYXBwbHlUbzoKICAgICAgLSAiaW5zdGFsbCIKICAgICAgLSAidXBncmFkZSIKICAgIHNlbnNpdGl2ZTogdHJ1ZQogIC0gbmFtZTogbXktc2Vjb25kLW91dHB1dAogICAgZGVzY3JpcHRpb246ICJNeSBzZWNvbmQgb3V0cHV0IgogICAgdHlwZTogYm9vbGVhbgogICAgc2Vuc2l0aXZlOiBmYWxzZQogIC0gbmFtZToga3ViZWNvbmZpZwogICAgdHlwZTogZmlsZQogICAgcGF0aDogL3Jvb3QvLmt1YmUvY29uZmlnCgptaXhpbnM6CiAgLSBleGVjCgppbnN0YWxsOgogIC0gZXhlYzoKICAgICAgZGVzY3JpcHRpb246ICJJbnN0YWxsIEhlbGxvIFdvcmxkIgogICAgICBjb21tYW5kOiBiYXNoCiAgICAgIGZsYWdzOgogICAgICAgIGM6IGVjaG8gSGVsbG8gV29ybGQKCnVwZ3JhZGU6CiAgLSBleGVjOgogICAgICBkZXNjcmlwdGlvbjogIldvcmxkIDIuMCIKICAgICAgY29tbWFuZDogYmFzaAogICAgICBmbGFnczoKICAgICAgICBjOiBlY2hvIFdvcmxkIDIuMAoKem9tYmllczoKICAtIGV4ZWM6CiAgICAgIGRlc2NyaXB0aW9uOiAiVHJpZ2dlciB6b21iaWUgYXBvY2FseXBzZSIKICAgICAgY29tbWFuZDogYmFzaAogICAgICBmbGFnczoKICAgICAgICBjOiBlY2hvIG9oIG5vZXMgbXkgYnJhaW5zCgp1bmluc3RhbGw6CiAgLSBleGVjOgogICAgICBkZXNjcmlwdGlvbjogIlVuaW5zdGFsbCBIZWxsbyBXb3JsZCIKICAgICAgY29tbWFuZDogYmFzaAogICAgICBmbGFnczoKICAgICAgICBjOiBlY2hvIEdvb2RieWUgV29ybGQK",
					},
				},
			},
			tag:                 kahn1dot01,
			shouldCacheManifest: true,
		},
		{
			name:   "no embedded manifest",
			tag:    kahnlatest,
			bundle: bundle.Bundle{},
		},
	}

	c := New(cfg.Config)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cb, err := c.StoreBundle(test.tag, test.bundle, nil)
			require.NoError(t, err, "StoreBundle failed")

			cachedManifestExists, _ := cfg.FileSystem.Exists(cb.BuildManifestPath())
			if test.shouldCacheManifest {
				assert.Equal(t, cb.BuildManifestPath(), cb.ManifestPath, "CachedBundle.ManifestPath was not set")
				assert.True(t, cachedManifestExists, "Expected the porter.yaml manifest to be cached but it wasn't")
			} else {
				assert.Empty(t, cb.ManifestPath, "CachedBundle.ManifestPath should not be set for non-porter bundles")
				assert.False(t, cachedManifestExists, "Expected no porter.yaml manifest to be cached but one was cached anyway. Not sure what happened there...")
			}
		})
	}
}
