package porter

import (
	"context"
	"errors"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/tests"
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
	err := opts.Validate(p.Config)
	require.NoError(t, err, "validating should not have failed")
}

func TestPublish_Validate_PorterYamlDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := PublishOptions{}
	err := opts.Validate(p.Config)
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
	err := opts.Validate(p.Config)
	assert.EqualError(t, err, "unable to access --archive mybuns.tgz: open /mybuns.tgz: file does not exist")

	p.FileSystem.WriteFile("mybuns.tgz", []byte("mybuns"), pkg.FileModeWritable)
	err = opts.Validate(p.Config)
	assert.EqualError(t, err, "must provide a value for --reference of the form REGISTRY/bundle:tag")

	opts.Reference = "myreg/mybuns:v0.1.0"
	err = opts.Validate(p.Config)
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
		assert.Equal(t, "example.com/neworg/apache:porter-83e8daf2fa98c1232fd8477a16eb8d0c", newInvImgName.String())
	})

	t.Run("has registry and org, bundle tag has subdomain", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/myorg/apache-installer", "example.com/neworg/bundles/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/bundles/apache:porter-83e8daf2fa98c1232fd8477a16eb8d0c", newInvImgName.String())
	})

	t.Run("has registry, org and subdomain, bundle tag has subdomain", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/myorg/myimgs/apache-installer", "example.com/neworg/bundles/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/bundles/apache:porter-e18bca98afc244c5d7a568be2cf6885f", newInvImgName.String())
	})

	t.Run("has registry, no org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("localhost:5000/apache-installer", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/apache:porter-2125d4f796f345561b13ec13a1f08e2d", newInvImgName.String())
	})

	t.Run("no registry, has org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("myorg/apache-installer", "example.com/anotherorg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/anotherorg/apache:porter-05885277937850e552535b74f7fc28a5", newInvImgName.String())
	})

	t.Run("org repeated in registry name", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("getporter/whalesayd", "getporter.azurecr.io/neworg/whalegap:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "getporter.azurecr.io/neworg/whalegap:porter-5cfeb864c54c7211a83a7d2ec5caaeb1", newInvImgName.String())
	})

	t.Run("org repeated in image name", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("getporter/getporter-hello-installer", "test.azurecr.io/neworg/hello:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "test.azurecr.io/neworg/hello:porter-5f484237ec91b98a63dd55846fb317ef", newInvImgName.String())
	})

	t.Run("src has no org, dst has no org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("apache", "example.com/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/apache:porter-b6efd606d118d0f62066e31419ff04cc", newInvImgName.String())
	})

	t.Run("src has no org, dst has org", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("apache", "example.com/neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "example.com/neworg/apache:porter-b6efd606d118d0f62066e31419ff04cc", newInvImgName.String())
	})

	t.Run("src has registry, dst has no registry (implicit docker.io)", func(t *testing.T) {
		newInvImgName, err := getNewImageNameFromBundleReference("oldregistry.com/apache", "neworg/apache:v0.1.0")
		require.NoError(t, err, "getNewImageNameFromBundleReference failed")
		assert.Equal(t, "docker.io/neworg/apache:porter-e8d04c0fd60dc2f793d2a865b899ca64", newInvImgName.String())
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

func TestPublish_ForceOverwrite(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		exists  bool
		force   bool
		wantErr string
	}{
		{name: "bundle doesn't exist, force not set", exists: false, force: false, wantErr: ""},
		{name: "bundle exists, force not set", exists: true, force: false, wantErr: "already exists in the destination registry"},
		{name: "bundle exists, force set", exists: true, force: true},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			p := NewTestPorter(t)
			defer p.Close()

			// Set up that the destination already exists
			p.TestRegistry.MockGetBundleMetadata = func(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) (cnabtooci.BundleMetadata, error) {
				if tc.exists {
					return cnabtooci.BundleMetadata{}, nil
				}
				return cnabtooci.BundleMetadata{}, cnabtooci.ErrNotFound{Reference: ref}
			}

			p.TestConfig.TestContext.AddTestDirectoryFromRoot("tests/testdata/mybuns", p.BundleDir)

			opts := PublishOptions{}
			opts.Force = tc.force

			err := opts.Validate(p.Config)
			require.NoError(t, err)

			err = p.Publish(ctx, opts)

			if tc.wantErr == "" {
				require.NoError(t, err, "Publish failed")
			} else {
				tests.RequireErrorContains(t, err, tc.wantErr)
			}
		})
	}
}
