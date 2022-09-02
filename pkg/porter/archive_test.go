package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/cnabio/image-relocation/pkg/image"
	"github.com/stretchr/testify/require"
)

func TestArchive_ParentDirDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := ArchiveOptions{}
	opts.Reference = "myreg/mybuns:v0.1.0"

	err := opts.Validate(context.Background(), []string{"/path/to/file"}, p.Porter)
	require.NoError(t, err, "expected no validation error to occur")

	err = p.Archive(context.Background(), opts)
	require.EqualError(t, err, "parent directory \"/path/to\" does not exist")
}

func TestArchive_Validate(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	testcases := []struct {
		name      string
		args      []string
		reference string
		wantError string
	}{
		{"no arg", nil, "", "destination file is required"},
		{"no tag", []string{"/path/to/file"}, "", "must provide a value for --reference of the form REGISTRY/bundle:tag"},
		{"too many args", []string{"/path/to/file", "moar args!"}, "myreg/mybuns:v0.1.0", "only one positional argument may be specified, the archive file name, but multiple were received: [/path/to/file moar args!]"},
		{"just right", []string{"/path/to/file"}, "myreg/mybuns:v0.1.0", ""},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := ArchiveOptions{}
			opts.Reference = tc.reference

			err := opts.Validate(context.Background(), tc.args, p.Porter)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err, "expected no validation error to occur")
			}
		})
	}
}

func TestArchive_ArchiveDirectory(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()
	ex := exporter{
		fs: p.FileSystem,
	}

	dir, err := ex.createArchiveFolder("examples/test-bundle-0.2.0")
	require.NoError(t, err)
	require.Contains(t, dir, "examples-test-bundle-0.2.0")

	tests.AssertDirectoryPermissionsEqual(t, dir, pkg.FileModeDirectory)
}

func TestArchive_AddImage(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	testcases := []struct {
		name           string
		relocationMap  relocation.ImageRelocationMap
		inputImg       string
		expectedImg    string
		hasErr         bool
		expectedErrMsg string
	}{
		{"no relocation map set", nil, "image:v0.1.0", "", true, "relocation map is not provided"},
		{"image not found in relocation map", relocation.ImageRelocationMap{"image:v0.1.0": "image@sha256:123"}, "not-found-image:v0.2.0", "", true, "can not locate the referenced image"},
		{"image successfully added", relocation.ImageRelocationMap{"image:v0.1.0": "image@sha256:123"}, "image:v0.1.0", "image@sha256:123", false, ""},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			baseImage := bundle.BaseImage{Image: tc.inputImg, Digest: "digest"}
			ex := exporter{relocationMap: tc.relocationMap, imageStore: mockImageStore{t: t, expected: tc.expectedImg}}
			err := ex.addImage(baseImage)
			if tc.hasErr {
				tests.RequireErrorContains(t, err, tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}

}

func TestArchive_PrepareArtifacts_Sorting(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	testcases := []struct {
		name          string
		relocationMap relocation.ImageRelocationMap
		inputImgs     []string
		expectedImgs  []string
	}{
		{"images sorted", relocation.ImageRelocationMap{"c:v0.1.0": "c@sha256:789", "a:v0.1.0": "a@sha256:123", "b:v0.1.0": "b@sha256:456"},
			[]string{"b:v0.1.0", "c:v0.1.0", "a:v0.1.0"},
			[]string{"a@sha256:123", "b@sha256:456", "c@sha256:789"}},
		{"numbers too", relocation.ImageRelocationMap{"a:v0.1.0": "a@sha256:123", "0b:v0.1.0": "0b@sha256:456"},
			[]string{"0b:v0.1.0", "a:v0.1.0"},
			[]string{"0b@sha256:456", "a@sha256:123"}},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			images := make(map[string]bundle.Image)
			b := cnab.NewBundle(bundle.Bundle{Images: images})
			for _, inputImg := range tc.inputImgs {
				images[inputImg] = bundle.Image{BaseImage: bundle.BaseImage{Image: inputImg, Digest: "digest"}}
			}
			collectedImages := make([]string, 0)
			imageStore := mockCollectingImageStore{t: t, addedImages: &collectedImages}
			ex := exporter{relocationMap: tc.relocationMap, imageStore: imageStore}

			err := ex.prepareArtifacts(b)

			require.Equal(t, tc.expectedImgs, collectedImages)
			require.NoError(t, err)
		})
	}

}

type mockCollectingImageStore struct {
	t           *testing.T
	addedImages *[]string
}

func (m mockCollectingImageStore) Add(img string) (contentDigest string, err error) {
	*m.addedImages = append(*m.addedImages, img)
	return "digest", nil
}

func (m mockCollectingImageStore) Push(dig image.Digest, src image.Name, dst image.Name) error {
	return nil
}

type mockImageStore struct {
	t        *testing.T
	expected string
}

func (m mockImageStore) Add(img string) (contentDigest string, err error) {
	require.Equal(m.t, m.expected, img)
	return "digest", nil
}

func (m mockImageStore) Push(dig image.Digest, src image.Name, dst image.Name) error {
	return nil
}
