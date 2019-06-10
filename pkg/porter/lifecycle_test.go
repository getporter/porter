package porter

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundlePullUpdateOpts_bundleCached(t *testing.T) {

	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	home, err := p.TestConfig.GetHomeDir()
	t.Logf("home dir is: %s", home)
	cacheDir, err := p.Cache.GetCacheDir()
	require.NoError(t, err, "should have had a porter cache dir")
	t.Logf("cache dir is: %s", cacheDir)
	p.TestConfig.TestContext.AddTestDirectory("testdata/cache", cacheDir)
	fullPath := filepath.Join(cacheDir, "887e7e65e39277f8744bd00278760b06/cnab/bundle.json")
	fileExists, err := p.TestConfig.TestContext.FileSystem.Exists(fullPath)
	require.True(t, fileExists, "this test requires that the file exist")

	cache := mockCache{
		findBundleMock: func(tag string) (string, bool, error) {
			return fullPath, true, nil
		},
	}
	p.Porter.Cache = &cache
	_, ok, err := p.Cache.FindBundle("deislabs/kubekahn:1.0")
	assert.True(t, ok, "should have found the bundle...")
	b := &BundleLifecycleOpts{
		BundlePullOptions: BundlePullOptions{
			Tag: "deislabs/kubekahn:1.0",
		},
	}
	err = b.populateOptsFromBundlePull(p.Porter)
	assert.NoError(t, err, "pulling bundle should not have resulted in an error")
	assert.Equal(t, "mysql", b.Name, "name should have matched testdata bundle")
	assert.Equal(t, fullPath, b.CNABFile, "the prepare method should have set the file to the fullpath")
}

func TestBundlePullUpdateOpts_pullError(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	cache := mockCache{
		findBundleMock: func(tag string) (string, bool, error) {
			return "", false, nil
		},
	}
	p.Porter.Cache = &cache

	b := &BundleLifecycleOpts{
		BundlePullOptions: BundlePullOptions{
			Tag: "deislabs/kubekahn:latest",
		},
	}
	err := b.populateOptsFromBundlePull(p.Porter)
	assert.Error(t, err, "pulling bundle should have resulted in an error")
	assert.Contains(t, err.Error(), "unable to pull bundle deislabs/kubekahn:latest")

}

func TestBundlePullUpdateOpts_cacheLies(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	cache := mockCache{
		findBundleMock: func(tag string) (string, bool, error) {
			return "/opt/not/here/bundle.json", true, nil
		},
	}
	p.Porter.Cache = &cache
	b := &BundleLifecycleOpts{
		BundlePullOptions: BundlePullOptions{
			Tag: "deislabs/kubekahn:latest",
		},
	}
	err := b.populateOptsFromBundlePull(p.Porter)
	assert.Error(t, err, "pulling bundle should have resulted in an error")
	assert.Contains(t, err.Error(), "unable to open bundle file")

}
