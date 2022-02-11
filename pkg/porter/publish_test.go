package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/cnab"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish_Validate_PorterYamlExists(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
	opts := PublishOptions{}
	err := opts.Validate(p.Context)
	require.NoError(t, err, "validating should not have failed")
}

func TestPublish_Validate_PorterYamlDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

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
	defer p.Teardown()

	opts := PublishOptions{
		ArchiveFile: "mybuns.tgz",
	}
	err := opts.Validate(p.Context)
	assert.EqualError(t, err, "unable to access --archive mybuns.tgz: open /mybuns.tgz: file does not exist")

	p.FileSystem.WriteFile("mybuns.tgz", []byte("mybuns"), 0600)
	err = opts.Validate(p.Context)
	assert.EqualError(t, err, "must provide a value for --reference of the form REGISTRY/bundle:tag")

	opts.Reference = "myreg/mybuns:v0.1.0"
	err = opts.Validate(p.Context)
	require.NoError(t, err, "validating should not have failed")
}

func TestPublish_validateTag(t *testing.T) {
	t.Run("tag is a Docker tag", func(t *testing.T) {
		opts := PublishOptions{
			Tag: "latest",
		}
		err := opts.validateTag()
		assert.NoError(t, err)
	})

	t.Run("tag is a full bundle reference with '@'", func(t *testing.T) {
		opts := PublishOptions{
			Tag: "myregistry.com/mybuns:v0.1.0",
		}
		err := opts.validateTag()
		assert.EqualError(t, err, "the --tag flag has been updated to designate just the Docker tag portion of the bundle reference; use --reference for the full bundle reference instead")
	})

	t.Run("tag is a full bundle reference with ':'", func(t *testing.T) {
		opts := PublishOptions{
			Tag: "myregistry.com/mybuns@abcde1234",
		}
		err := opts.validateTag()
		assert.EqualError(t, err, "the --tag flag has been updated to designate just the Docker tag portion of the bundle reference; use --reference for the full bundle reference instead")
	})
}

func TestPublish_getNewImageNameFromBundleReference(t *testing.T) {
	t.Run("has registry and org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/myorg/apache-installer", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/apache:3a7b06390b4421c38c57e5e44475f1ba", newInvImgName.String())
	})

	t.Run("has registry and org, bundle tag has subdomain", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/myorg/apache-installer", "example.com/neworg/bundles/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/bundles/apache:9d56a9aebe9641c667a34b04be1d259e", newInvImgName.String())
	})

	t.Run("has registry, org and subdomain, bundle tag has subdomain", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/myorg/myimgs/apache-installer", "example.com/neworg/bundles/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/bundles/apache:9d56a9aebe9641c667a34b04be1d259e", newInvImgName.String())
	})

	t.Run("has registry, no org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/apache-installer", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/apache:3a7b06390b4421c38c57e5e44475f1ba", newInvImgName.String())
	})

	t.Run("no registry, has org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("myorg/apache-installer", "example.com/anotherorg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/anotherorg/apache:d0bbfdcdda80ded171de26cc7bb55fe5", newInvImgName.String())
	})

	t.Run("org repeated in registry name", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("getporter/whalesayd", "getporter.azurecr.io/neworg/whalegap:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "getporter.azurecr.io/neworg/whalegap:497b3e01ed1153e542a19df184607866", newInvImgName.String())
	})

	t.Run("org repeated in image name", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("getporter/getporter-hello-installer", "test.azurecr.io/neworg/hello:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "test.azurecr.io/neworg/hello:1876516fb3650bdaf1f6b764522b720e", newInvImgName.String())
	})

	t.Run("src has no org, dst has no org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("apache", "example.com/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/apache:c85c90de79df5b0bbdb39cb06cecc53f", newInvImgName.String())
	})

	t.Run("src has no org, dst has org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("apache", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/apache:da232bb13d5b6ab0214f720968992e25", newInvImgName.String())
	})

	t.Run("src has registry, dst has no registry (implicit docker.io)", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("oldregistry.com/apache", "neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "docker.io/neworg/apache:69772e4894006c2ebdee3e4ccd6ee718", newInvImgName.String())
	})
}

func TestPublish_UpdateBundleWithNewImage(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	bun := cnab.ExtendedBundle{bundle.Bundle{
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
	}}
	tag := "myneworg/mynewbuns"

	digest, err := image.NewDigest("sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687")
	require.NoError(t, err, "should have successfully created a digest")

	// update invocation image
	newInvImgName, err := getNewImageNameFromBundleReference(bun.InvocationImages[0].Image, tag)
	require.NoError(t, err, "should have successfully derived new image name from bundle tag")

	err = p.updateBundleWithNewImage(bun, newInvImgName, digest, 0)
	require.NoError(t, err, "updating bundle with new image should not have failed")
	require.Equal(t, tag+":2ebf8f866994ea333390db1ea281de90@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", bun.InvocationImages[0].Image)
	require.Equal(t, "sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", bun.InvocationImages[0].Digest)

	// update image
	newImgName, err := getNewImageNameFromBundleReference(bun.Images["myimg"].Image, tag)
	require.NoError(t, err, "should have successfully derived new image name from bundle tag")

	err = p.updateBundleWithNewImage(bun, newImgName, digest, "myimg")
	require.NoError(t, err, "updating bundle with new image should not have failed")
	require.Equal(t, tag+":e7834f03677bd970b2a458c84d0cdff0@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", bun.Images["myimg"].Image)
	require.Equal(t, "sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", bun.Images["myimg"].Digest)
}

func TestPublish_RefreshCachedBundle(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	bundleRef := cnab.BundleReference{
		Reference:  cnab.MustParseOCIReference("myreg/mybuns"),
		Definition: cnab.ExtendedBundle{bundle.Bundle{Name: "myreg/mybuns"}},
	}

	// No-Op; bundle does not yet exist in cache
	err := p.refreshCachedBundle(bundleRef)
	require.NoError(t, err, "should have not errored out if bundle does not yet exist in cache")

	// Save bundle in cache
	cachedBundle, err := p.Cache.StoreBundle(bundleRef)
	require.NoError(t, err, "should have successfully stored bundle")

	// Get file mod time
	file, err := p.FileSystem.Stat(cachedBundle.BundlePath)
	require.NoError(t, err)
	origBunPathTime := file.ModTime()

	// Should refresh cache
	err = p.refreshCachedBundle(bundleRef)
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
	defer p.Teardown()

	bundleRef := cnab.BundleReference{
		Reference:  cnab.MustParseOCIReference("myreg/mybuns"),
		Definition: cnab.ExtendedBundle{bundle.Bundle{Name: "myreg/mybuns"}},
	}

	p.TestCache.FindBundleMock = func(ref cnab.OCIReference) (cachedBundle cache.CachedBundle, found bool, err error) {
		// force the bundle to be found
		return cache.CachedBundle{}, true, nil
	}
	p.TestCache.StoreBundleMock = func(bundleRef cnab.BundleReference) (cachedBundle cache.CachedBundle, err error) {
		// sabotage the bundle refresh
		return cache.CachedBundle{}, errors.New("error trying to store bundle")
	}

	err := p.refreshCachedBundle(bundleRef)
	require.NoError(t, err, "should have not errored out even if cache.StoreBundle does")

	gotStderr := p.TestConfig.TestContext.GetError()
	require.Equal(t, "warning: unable to update cache for bundle myreg/mybuns: error trying to store bundle\n", gotStderr)
}
