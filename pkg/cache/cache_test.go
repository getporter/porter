package cache

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	kahn1dot0Hash = "887e7e65e39277f8744bd00278760b06"
	kahn1dot01    = cnab.MustParseOCIReference("deislabs/kubekahn:1.0")
	kahnlatest    = cnab.MustParseOCIReference("deislabs/kubekahn:latest")
)

func TestFindBundleCacheExists(t *testing.T) {
	t.Parallel()

	cfg := config.NewTestConfig(t)
	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)
	c := New(cfg.Config)

	_, ok, err := c.FindBundle(kahnlatest)
	assert.NoError(t, err, "the cache dir should exist, no error should have happened")
	assert.False(t, ok, "the bundle shouldn't exist")
}

func TestFindBundleCacheDoesNotExist(t *testing.T) {
	t.Parallel()

	cfg := config.NewTestConfig(t)
	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)
	c := New(cfg.Config)

	_, ok, err := c.FindBundle(kahnlatest)
	assert.NoError(t, err, "the cache dir doesn't exist, but this shouldn't be an error")
	assert.False(t, ok, "the bundle shouldn't exist")
}

func TestFindBundleBundleCached(t *testing.T) {
	t.Parallel()

	cfg := config.NewTestConfig(t)
	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)
	expectedCacheDirectory := filepath.Join(cacheDir, kahn1dot0Hash)
	expectedCacheCNABDirectory := filepath.Join(expectedCacheDirectory, "cnab")
	expectedCacheFile := filepath.Join(expectedCacheCNABDirectory, "bundle.json")
	foundIt, err := cfg.Config.FileSystem.Exists(expectedCacheFile)
	require.NoError(t, err, "the cache dir should exist, no error should have happened")
	require.True(t, foundIt, "test data not loaded")
	c := New(cfg.Config)

	cb, ok, err := c.FindBundle(kahn1dot01)
	require.NoError(t, err, "the cache dir should exist, no error should have happened")
	require.True(t, ok, "the bundle should exist")
	assert.Equal(t, expectedCacheFile, cb.BundlePath)
}

func TestFindBundleBundleNotCached(t *testing.T) {
	t.Parallel()

	cfg := config.NewTestConfig(t)
	c := New(cfg.Config)
	cb, ok, err := c.FindBundle(kahnlatest)
	require.NoError(t, err, "the cache dir should exist, no error should have happened")
	assert.False(t, ok, "the bundle should not exist")
	assert.Empty(t, cb.BundlePath, "should not have a path")
}

func TestCacheWriteNoCacheDir(t *testing.T) {
	t.Parallel()

	cfg := config.NewTestConfig(t)
	cfg.TestContext.AddTestFile("testdata/cnab/bundle.json", "/cnab/bundle.json")
	bun, err := cnab.LoadBundle(cfg.Context, "/cnab/bundle.json")
	require.NoError(t, err, "bundle should have been valid")

	c := New(cfg.Config)
	cb, err := c.StoreBundle(cnab.BundleReference{Reference: kahn1dot01, Definition: bun})
	assert.NoError(t, err, "storing bundle should have succeeded")

	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	expectedCacheDirectory := filepath.Join(cacheDir, kahn1dot0Hash)
	expectedCacheCNABDirectory := filepath.Join(expectedCacheDirectory, "cnab")
	expectedCacheFile := filepath.Join(expectedCacheCNABDirectory, "bundle.json")

	assert.Equal(t, expectedCacheFile, cb.BundlePath)
}

func TestCacheWriteCacheDirExists(t *testing.T) {
	t.Parallel()

	cfg := config.NewTestConfig(t)
	home, err := cfg.Config.GetHomeDir()
	require.NoError(t, err, "should have had a porter home dir")
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestFile("testdata/cnab/bundle.json", "/cnab/bundle.json")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)
	bun, err := cnab.LoadBundle(cfg.Context, "/cnab/bundle.json")
	require.NoError(t, err, "bundle should have been valid")

	c := New(cfg.Config)
	var reloMap relocation.ImageRelocationMap
	bundleRef := cnab.BundleReference{
		Reference:     kahn1dot01,
		Definition:    bun,
		RelocationMap: reloMap,
		Digest:        digest.Digest("sha256:2249472f86d0cea9ac8809331931e9100e1d0464afff3d2869bbb8dedfe2d396"),
	}
	cb, err := c.StoreBundle(bundleRef)

	expectedCacheDirectory := filepath.Join(cacheDir, kahn1dot0Hash)
	expectedCacheCNABDirectory := filepath.Join(expectedCacheDirectory, "cnab")
	expectedCacheFile := filepath.Join(expectedCacheCNABDirectory, "bundle.json")

	assert.Equal(t, expectedCacheFile, cb.BundlePath)
	assert.NoError(t, err, "storing bundle should have succeeded")

	var meta Metadata
	expectedMetaFile := filepath.Join(expectedCacheDirectory, "metadata.json")
	require.NoError(t, encoding.UnmarshalFile(cfg.FileSystem, expectedMetaFile, &meta))
	assert.Equal(t, Metadata{Reference: bundleRef.Reference, Digest: bundleRef.Digest}, meta, "incorrect metadata.json persisted")
}

func TestStoreRelocationMapping(t *testing.T) {
	tests := []struct {
		name              string
		relocationMapping relocation.ImageRelocationMap
		tag               cnab.OCIReference
		bundle            cnab.ExtendedBundle
		wantedReloPath    string
		err               error
	}{
		{
			name:   "relocation file gets a path",
			bundle: cnab.ExtendedBundle{},
			tag:    kahn1dot01,
			relocationMapping: relocation.ImageRelocationMap{
				"asd": "asdf",
			},
			wantedReloPath: path.Join("/.porter/cache", kahn1dot0Hash, "cnab", "relocation-mapping.json"),
		},
		{
			name:           "no relocation file gets no path",
			tag:            kahnlatest,
			bundle:         cnab.ExtendedBundle{},
			wantedReloPath: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			cfg := config.NewTestConfig(t)
			c := New(cfg.Config)

			cb, err := c.StoreBundle(cnab.BundleReference{Reference: tc.tag, Definition: tc.bundle, RelocationMap: tc.relocationMapping})
			assert.NoError(t, err, fmt.Sprintf("didn't expect storage error for test %s", tc.name))
			assert.Equal(t, tc.wantedReloPath, cb.RelocationFilePath, "didn't get expected path for store")

			cb, _, err = c.FindBundle(tc.tag)
			assert.NoError(t, err, fmt.Sprintf("didn't expect find bundle error for test %s", tc.tag))
			assert.Equal(t, tc.wantedReloPath, cb.RelocationFilePath, "didn't get expected path for load")
		})
	}
}

func TestStoreManifest(t *testing.T) {
	tests := []struct {
		name                string
		tag                 cnab.OCIReference
		bundle              cnab.ExtendedBundle
		shouldCacheManifest bool
		err                 error
	}{
		{
			name: "embedded manifest",
			bundle: cnab.NewBundle(bundle.Bundle{
				Custom: map[string]interface{}{
					"sh.porter": map[string]interface{}{
						"manifest": "bmFtZTogSEVMTE9fQ1VTVE9NCnZlcnNpb246IDAuMS4wCmRlc2NyaXB0aW9uOiAiQSBidW5kbGUgd2l0aCBhIGN1c3RvbSBhY3Rpb24iCnRhZzogZ2V0cG9ydGVyL3BvcnRlci1oZWxsbzp2MC4xLjAKaW52b2NhdGlvbkltYWdlOiBnZXRwb3J0ZXIvcG9ydGVyLWhlbGxvLWluc3RhbGxlcjowLjEuMAoKY3JlZGVudGlhbHM6CiAgLSBuYW1lOiBteS1maXJzdC1jcmVkCiAgICBlbnY6IE1ZX0ZJUlNUX0NSRUQKICAtIG5hbWU6IG15LXNlY29uZC1jcmVkCiAgICBkZXNjcmlwdGlvbjogIk15IHNlY29uZCBjcmVkIgogICAgcGF0aDogL3BhdGgvdG8vbXktc2Vjb25kLWNyZWQKCmltYWdlczogCiAgIHNvbWV0aGluZzoKICAgICAgZGVzY3JpcHRpb246ICJhbiBpbWFnZSIKICAgICAgaW1hZ2VUeXBlOiAiZG9ja2VyIgogICAgICByZXBvc2l0b3J5OiAiZ2V0cG9ydGVyL2JvbyIKCnBhcmFtZXRlcnM6CiAgLSBuYW1lOiBteS1maXJzdC1wYXJhbQogICAgdHlwZTogaW50ZWdlcgogICAgZGVmYXVsdDogOQogICAgZW52OiBNWV9GSVJTVF9QQVJBTQogICAgYXBwbHlUbzoKICAgICAgLSAiaW5zdGFsbCIKICAtIG5hbWU6IG15LXNlY29uZC1wYXJhbQogICAgZGVzY3JpcHRpb246ICJNeSBzZWNvbmQgcGFyYW1ldGVyIgogICAgdHlwZTogc3RyaW5nCiAgICBkZWZhdWx0OiBzcHJpbmctbXVzaWMtZGVtbwogICAgcGF0aDogL3BhdGgvdG8vbXktc2Vjb25kLXBhcmFtCiAgICBzZW5zaXRpdmU6IHRydWUKCm91dHB1dHM6CiAgLSBuYW1lOiBteS1maXJzdC1vdXRwdXQKICAgIHR5cGU6IHN0cmluZwogICAgYXBwbHlUbzoKICAgICAgLSAiaW5zdGFsbCIKICAgICAgLSAidXBncmFkZSIKICAgIHNlbnNpdGl2ZTogdHJ1ZQogIC0gbmFtZTogbXktc2Vjb25kLW91dHB1dAogICAgZGVzY3JpcHRpb246ICJNeSBzZWNvbmQgb3V0cHV0IgogICAgdHlwZTogYm9vbGVhbgogICAgc2Vuc2l0aXZlOiBmYWxzZQogIC0gbmFtZToga3ViZWNvbmZpZwogICAgdHlwZTogZmlsZQogICAgcGF0aDogL3Jvb3QvLmt1YmUvY29uZmlnCgptaXhpbnM6CiAgLSBleGVjCgppbnN0YWxsOgogIC0gZXhlYzoKICAgICAgZGVzY3JpcHRpb246ICJJbnN0YWxsIEhlbGxvIFdvcmxkIgogICAgICBjb21tYW5kOiBiYXNoCiAgICAgIGZsYWdzOgogICAgICAgIGM6IGVjaG8gSGVsbG8gV29ybGQKCnVwZ3JhZGU6CiAgLSBleGVjOgogICAgICBkZXNjcmlwdGlvbjogIldvcmxkIDIuMCIKICAgICAgY29tbWFuZDogYmFzaAogICAgICBmbGFnczoKICAgICAgICBjOiBlY2hvIFdvcmxkIDIuMAoKem9tYmllczoKICAtIGV4ZWM6CiAgICAgIGRlc2NyaXB0aW9uOiAiVHJpZ2dlciB6b21iaWUgYXBvY2FseXBzZSIKICAgICAgY29tbWFuZDogYmFzaAogICAgICBmbGFnczoKICAgICAgICBjOiBlY2hvIG9oIG5vZXMgbXkgYnJhaW5zCgp1bmluc3RhbGw6CiAgLSBleGVjOgogICAgICBkZXNjcmlwdGlvbjogIlVuaW5zdGFsbCBIZWxsbyBXb3JsZCIKICAgICAgY29tbWFuZDogYmFzaAogICAgICBmbGFnczoKICAgICAgICBjOiBlY2hvIEdvb2RieWUgV29ybGQK",
					},
				},
			}),
			tag:                 kahn1dot01,
			shouldCacheManifest: true,
		},
		{
			name: "porter stamp, no manifest",
			bundle: cnab.NewBundle(bundle.Bundle{
				Custom: map[string]interface{}{
					"sh.porter": map[string]interface{}{
						"manifestDigest": "abc123",
					},
				},
			}),
			tag:                 kahn1dot01,
			shouldCacheManifest: false,
		},
		{
			name:   "no embedded manifest",
			tag:    kahnlatest,
			bundle: cnab.ExtendedBundle{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			cfg := config.NewTestConfig(t)
			home, _ := cfg.Config.GetHomeDir()
			cacheDir := filepath.Join(home, "cache")
			cfg.TestContext.AddTestDirectory("testdata", cacheDir)
			c := New(cfg.Config)

			cb, err := c.StoreBundle(cnab.BundleReference{Reference: tc.tag, Definition: tc.bundle})
			require.NoError(t, err, "StoreBundle failed")

			cachedManifestExists, _ := cfg.FileSystem.Exists(cb.BuildManifestPath())
			if tc.shouldCacheManifest {
				assert.Equal(t, cb.BuildManifestPath(), cb.ManifestPath, "CachedBundle.ManifestPath was not set")
				assert.True(t, cachedManifestExists, "Expected the porter.yaml manifest to be cached but it wasn't")
			} else {
				assert.Empty(t, cb.ManifestPath, "CachedBundle.ManifestPath should not be set for non-porter bundles")
				assert.False(t, cachedManifestExists, "Expected porter.yaml manifest to not be cached but one was cached anyway. Not sure what happened there...")
			}
		})
	}
}

func TestCache_StoreBundle_Overwrite(t *testing.T) {
	t.Parallel()

	cfg := config.NewTestConfig(t)
	home, _ := cfg.Config.GetHomeDir()
	cacheDir := filepath.Join(home, "cache")
	cfg.TestContext.AddTestDirectory("testdata", cacheDir)
	c := New(cfg.Config)

	// Setup an existing bundle with some extraneous junk that would not
	// be overwritten
	cb := CachedBundle{BundleReference: cnab.BundleReference{Reference: kahn1dot01}}
	cb.SetCacheDir(cacheDir)
	cfg.FileSystem.Create(cb.BuildManifestPath())
	cfg.FileSystem.Create(cb.BuildRelocationFilePath())
	junkPath := filepath.Join(cb.cacheDir, "junk.txt")
	cfg.FileSystem.Create(junkPath)

	// Refresh the cache
	cb, err := c.StoreBundle(cb.BundleReference)
	require.NoError(t, err, "StoreBundle failed")

	exists, _ := cfg.FileSystem.Exists(cb.BuildBundlePath())
	assert.True(t, exists, "bundle.json should have been written in the refreshed cache")

	exists, _ = cfg.FileSystem.Exists(cb.BuildManifestPath())
	assert.False(t, exists, "porter.yaml should have been deleted from the bundle cache")

	exists, _ = cfg.FileSystem.Exists(cb.BuildRelocationFilePath())
	assert.False(t, exists, "relocation-mapping.json should have been deleted from the bundle cache")

	exists, _ = cfg.FileSystem.Exists(junkPath)
	assert.False(t, exists, "the random file should have been deleted from the bundle cache")
}
