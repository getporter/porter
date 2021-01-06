package porter

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/cache"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish_Validate_PorterYamlExists(t *testing.T) {
	p := NewTestPorter(t)

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
	opts := PublishOptions{}
	err := opts.Validate(p.Context)
	require.NoError(t, err, "validating should not have failed")
}

func TestPublish_Validate_PorterYamlDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)

	opts := PublishOptions{}
	err := opts.Validate(p.Context)
	require.Error(t, err, "validation should have failed")
	assert.EqualError(
		t,
		err,
		"could not find porter.yaml in the current directory, make sure you are in the right directory or specify the porter manifest with --file",
		"porter.yaml not present so should have failed validation",
	)
}

func TestPublish_Validate_ArchivePath(t *testing.T) {
	p := NewTestPorter(t)

	opts := PublishOptions{
		ArchiveFile: "mybuns.tgz",
	}
	err := opts.Validate(p.Context)
	require.Error(t, err, "expected Validate to fail")
	assert.True(t, os.IsNotExist(errors.Cause(err)), "expected a file not found error")

	p.FileSystem.WriteFile("mybuns.tgz", []byte("mybuns"), os.ModePerm)
	err = opts.Validate(p.Context)
	assert.EqualError(t, err, "must provide a value for --tag of the form REGISTRY/bundle:tag")

	opts.Tag = "myreg/mybuns:v0.1.0"
	err = opts.Validate(p.Context)
	require.NoError(t, err, "validating should not have failed")
}

func TestPublish_getNewImageNameFromBundleTag(t *testing.T) {
	t.Run("has registry and org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("localhost:5000/myorg/apache-installer", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "example.com/neworg/apache-installer", newInvImgName.String())
	})

	t.Run("has registry and org, bundle tag has subdomain", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("localhost:5000/myorg/apache-installer", "example.com/neworg/bundles/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "example.com/neworg/bundles/apache-installer", newInvImgName.String())
	})

	t.Run("has registry, org and subdomain, bundle tag has subdomain", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("localhost:5000/myorg/myimgs/apache-installer", "example.com/neworg/bundles/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "example.com/neworg/bundles/apache-installer", newInvImgName.String())
	})

	t.Run("has registry, no org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("localhost:5000/apache-installer", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "example.com/neworg/apache-installer", newInvImgName.String())
	})

	t.Run("no registry, has org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("myorg/apache-installer", "example.com/anotherorg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "example.com/anotherorg/apache-installer", newInvImgName.String())
	})

	t.Run("org repeated in registry name", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("getporter/whalesayd", "getporter.azurecr.io/neworg/whalegap:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "getporter.azurecr.io/neworg/whalesayd", newInvImgName.String())
	})

	t.Run("org repeated in image name", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("getporter/getporter-hello-installer", "test.azurecr.io/neworg/hello:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "test.azurecr.io/neworg/getporter-hello-installer", newInvImgName.String())
	})

	t.Run("src has no org, dst has no org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("apache", "example.com/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "example.com/apache", newInvImgName.String())
	})

	t.Run("src has no org, dst has org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("apache", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "example.com/neworg/apache", newInvImgName.String())
	})

	t.Run("src has registry, dst has no registry (implicit docker.io)", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleTag("oldregistry.com/apache", "neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleTag failed")
		assert.Equal(t, "docker.io/neworg/apache", newInvImgName.String())
	})
}

func TestPublish_UpdateBundleWithNewImage(t *testing.T) {
	p := NewTestPorter(t)

	bun := bundle.Bundle{
		Name: "mybuns",
		InvocationImages: []bundle.InvocationImage{
			{
				BaseImage: bundle.BaseImage{
					Image:  "myorg/myinvimg",
					Digest: "abc",
				},
			},
		},
		Images: map[string]bundle.Image{
			"myimg": {
				BaseImage: bundle.BaseImage{
					Image:  "myorg/myimg",
					Digest: "abc",
				},
			},
		},
	}
	tag := "myneworg/mynewbuns"

	digest, err := image.NewDigest("sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687")
	require.NoError(t, err, "should have successfully created a digest")

	// update invocation image
	newInvImgName, err := getNewImageNameFromBundleTag(bun.InvocationImages[0].Image, tag)
	require.NoError(t, err, "should have successfully derived new image name from bundle tag")

	err = p.updateBundleWithNewImage(bun, newInvImgName, digest, 0)
	require.NoError(t, err, "updating bundle with new image should not have failed")
	require.Equal(t, "docker.io/myneworg/myinvimg@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", bun.InvocationImages[0].Image)
	require.Equal(t, "sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", bun.InvocationImages[0].Digest)

	// update image
	newImgName, err := getNewImageNameFromBundleTag(bun.Images["myimg"].Image, tag)
	require.NoError(t, err, "should have successfully derived new image name from bundle tag")

	err = p.updateBundleWithNewImage(bun, newImgName, digest, "myimg")
	require.NoError(t, err, "updating bundle with new image should not have failed")
	require.Equal(t, "docker.io/myneworg/myimg@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", bun.Images["myimg"].Image)
	require.Equal(t, "sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", bun.Images["myimg"].Digest)
}

func TestPublish_RefreshCachedBundle(t *testing.T) {
	p := NewTestPorter(t)

	bun := bundle.Bundle{Name: "myreg/mybuns"}
	tag := "myreg/mybuns"

	// No-Op; bundle does not yet exist in cache
	err := p.refreshCachedBundle(bun, tag, nil)
	require.NoError(t, err, "should have not errored out if bundle does not yet exist in cache")

	// Save bundle in cache
	cachedBundle, err := p.Cache.StoreBundle(tag, bun, nil)
	require.NoError(t, err, "should have successfully stored bundle")

	// Get file mod time
	file, err := p.FileSystem.Stat(cachedBundle.BundlePath)
	require.NoError(t, err)
	origBunPathTime := file.ModTime()

	// Should refresh cache
	err = p.refreshCachedBundle(bun, tag, nil)
	require.NoError(t, err, "should have successfully updated the cache")

	// Get file mod time
	file, err = p.FileSystem.Stat(cachedBundle.BundlePath)
	require.NoError(t, err)
	updatedBunPathTime := file.ModTime()

	// Verify mod times differ
	require.NotEqual(t, updatedBunPathTime, origBunPathTime,
		"bundle.json file should have an updated mod time per cache refresh")
}

func TestPublish_RefreshCachedBundle_OnlyWarning(t *testing.T) {
	p := NewTestPorter(t)
	bun := bundle.Bundle{Name: "myreg/mybuns"}
	tag := "myreg/mybuns"

	p.TestCache.FindBundleMock = func(s string) (cachedBundle cache.CachedBundle, found bool, err error) {
		// force the bundle to be found
		return cache.CachedBundle{}, true, nil
	}
	p.TestCache.StoreBundleMock = func(s string, b bundle.Bundle, relocationMap *relocation.ImageRelocationMap) (cachedBundle cache.CachedBundle, err error) {
		// sabotage the bundle refresh
		return cache.CachedBundle{}, errors.New("error trying to store bundle")
	}

	err := p.refreshCachedBundle(bun, tag, nil)
	require.NoError(t, err, "should have not errored out even if cache.StoreBundle does")

	gotStderr := p.TestConfig.TestContext.GetError()
	require.Equal(t, "warning: unable to update cache for bundle myreg/mybuns: error trying to store bundle\n", gotStderr)
}
