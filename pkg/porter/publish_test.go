package porter

import (
	"errors"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/cnab"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/cnabio/image-relocation/pkg/image"
	"github.com/cnabio/image-relocation/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish_Validate_PorterYamlExists(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
	opts := PublishOptions{}
	err := opts.Validate(p.Context)
	require.NoError(t, err, "validating should not have failed")
}

func TestPublish_Validate_PorterYamlDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := PublishOptions{}
	err := opts.Validate(p.Context)
	require.Error(t, err, "validation should have failed")
	assert.EqualError(
		t,
		err,
		"could not find porter.yaml in the current directory /, make sure you are in the right directory or specify the porter manifest with --file",
		"porter.yaml not present so should have failed validation",
	)
}

func TestPublish_Validate_ArchivePath(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := PublishOptions{
		ArchiveFile: "mybuns.tgz",
	}
	err := opts.Validate(p.Context)
	assert.EqualError(t, err, "unable to access --archive mybuns.tgz: open /mybuns.tgz: file does not exist")

	p.FileSystem.WriteFile("mybuns.tgz", []byte("mybuns"), pkg.FileModeWritable)
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
		assert.Equal(t, "example.com/neworg/apache:eeff130ab77e7a1cb215717bde0a9b03", newInvImgName.String())
	})

	t.Run("has registry and org, bundle tag has subdomain", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/myorg/apache-installer", "example.com/neworg/bundles/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/bundles/apache:f394e29a474d79328491043a00d19b1b", newInvImgName.String())
	})

	t.Run("has registry, org and subdomain, bundle tag has subdomain", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/myorg/myimgs/apache-installer", "example.com/neworg/bundles/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/bundles/apache:5b055c1a5255d993df3faebcb5f1c682", newInvImgName.String())
	})

	t.Run("has registry, no org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/apache-installer", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/apache:20a9d1b7ba65580e24d1866a4ba3efbe", newInvImgName.String())
	})

	t.Run("no registry, has org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("myorg/apache-installer", "example.com/anotherorg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/anotherorg/apache:28e5c9d710b12e7ca785df72278f0287", newInvImgName.String())
	})

	t.Run("org repeated in registry name", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("getporter/whalesayd", "getporter.azurecr.io/neworg/whalegap:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "getporter.azurecr.io/neworg/whalegap:05af4cbacb54498b2440474276ff8cb1", newInvImgName.String())
	})

	t.Run("org repeated in image name", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("getporter/getporter-hello-installer", "test.azurecr.io/neworg/hello:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "test.azurecr.io/neworg/hello:f77402dbc280b34267626ce14a5dd843", newInvImgName.String())
	})

	t.Run("src has no org, dst has no org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("apache", "example.com/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/apache:f50507960a059fad3bd7f96bddcf10af", newInvImgName.String())
	})

	t.Run("src has no org, dst has org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("apache", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/apache:0c99b372d2f77a1424a733d625a1bbd9", newInvImgName.String())
	})

	t.Run("src has registry, dst has no registry (implicit docker.io)", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("oldregistry.com/apache", "neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "docker.io/neworg/apache:f5132d75619141add0c06d3d2640d3d3", newInvImgName.String())
	})
}

func TestPublish_RelocateImage(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	originImg := "myorg/myinvimg"
	tag := "myneworg/mynewbuns"
	digest, err := image.NewDigest("sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687")
	require.NoError(t, err, "should have successfully created a digest")

	testcases := []struct {
		description   string
		relocationMap relocation.ImageRelocationMap
		layout        registry.Layout
		wantErr       error
	}{
		{description: "has relocation mapping defined", relocationMap: relocation.ImageRelocationMap{"myorg/myinvimg": "private/myinvimg"}, layout: mockRegistryLayout{expectedDigest: digest}},
		{description: "empty relocation map", relocationMap: relocation.ImageRelocationMap{}, layout: mockRegistryLayout{expectedDigest: digest}},
		{description: "failed to update", relocationMap: relocation.ImageRelocationMap{"myorg/myinvimg": "private/myinvimg"}, layout: mockRegistryLayout{hasError: true}, wantErr: errors.New("unable to push updated image")},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			newMap, err := p.relocateImage(tc.relocationMap, tc.layout, originImg, tag)
			if tc.wantErr != nil {
				require.ErrorContains(t, err, tc.wantErr.Error())
				return
			}
			require.Equal(t, tag+"@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", newMap[originImg])
		})
	}
}

type mockRegistryLayout struct {
	hasError       bool
	expectedDigest image.Digest
}

func (m mockRegistryLayout) Add(name image.Name) (image.Digest, error) {
	return image.EmptyDigest, nil
}

func (m mockRegistryLayout) Push(digest image.Digest, name image.Name) error {
	if m.hasError {
		return errors.New("failed to add image")
	}
	return nil
}

func (m mockRegistryLayout) Find(n image.Name) (image.Digest, error) {
	return m.expectedDigest, nil
}

func TestPublish_RefreshCachedBundle(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	bundleRef := cnab.BundleReference{
		Reference:  cnab.MustParseOCIReference("myreg/mybuns"),
		Definition: cnab.NewBundle(bundle.Bundle{Name: "myreg/mybuns"}),
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
	defer p.Close()

	bundleRef := cnab.BundleReference{
		Reference:  cnab.MustParseOCIReference("myreg/mybuns"),
		Definition: cnab.NewBundle(bundle.Bundle{Name: "myreg/mybuns"}),
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

func TestPublish_RewriteImageWithDigest(t *testing.T) {
	// change from our temporary tag for the invocation image to using ONLY the digest
	p := NewTestPorter(t)
	defer p.Close()

	digestedImg, err := p.rewriteImageWithDigest("example/mybuns:temp-tag", "sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687")
	require.NoError(t, err)
	assert.Equal(t, "example/mybuns@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687", digestedImg)
}
