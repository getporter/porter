package porter

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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
		name             string
		args             []string
		reference        string
		compressionLevel string
		wantError        string
	}{
		{"no arg", nil, "", "", "destination file is required"},
		{"no tag", []string{"/path/to/file"}, "", "", "must provide a value for --reference of the form REGISTRY/bundle:tag"},
		{"too many args", []string{"/path/to/file", "moar args!"}, "myreg/mybuns:v0.1.0", "", "only one positional argument may be specified, the archive file name, but multiple were received: [/path/to/file moar args!]"},
		{"invalid compression level", []string{"/path/to/file"}, "myreg/mybuns:v0.1.0", "NotValidCompression", "invalid compression level: NotValidCompression"},
		{"no compression level", []string{"/path/to/file"}, "myreg/mybuns:v0.1.0", "NoCompression", ""},
		{"just right", []string{"/path/to/file"}, "myreg/mybuns:v0.1.0", "", ""},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := ArchiveOptions{}
			opts.Reference = tc.reference
			opts.CompressionLevel = tc.compressionLevel

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

// TestArchive_CustomTar_MetadataFirst verifies that bundle.json,
// relocation-mapping.json, and artifacts/layout/{oci-layout,index.json} all
// come before artifacts/layout/blobs/ in the tar stream, so a client
// streaming the archive can read the bundle manifest and resolve every
// image's digest (via index.json) without first reading through the
// (potentially many-gigabyte) blobs/ tree.
// See https://github.com/getporter/porter/issues/2197.
func TestArchive_CustomTar_MetadataFirst(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "bundle.json"), []byte(`{}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "relocation-mapping.json"), []byte(`{}`), 0644))

	layoutDir := filepath.Join(dir, "artifacts", "layout")
	require.NoError(t, os.MkdirAll(layoutDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(layoutDir, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(layoutDir, "index.json"), []byte(`{"schemaVersion":2,"manifests":[]}`), 0644))

	blobsDir := filepath.Join(layoutDir, "blobs", "sha256")
	require.NoError(t, os.MkdirAll(blobsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(blobsDir, "abc123"), []byte("blob"), 0644))

	ex := &exporter{}
	rc, err := ex.CustomTar(context.Background(), dir, gzip.DefaultCompression)
	require.NoError(t, err)
	defer rc.Close()

	gz, err := gzip.NewReader(rc)
	require.NoError(t, err)
	defer gz.Close()

	var names []string
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		names = append(names, hdr.Name)
	}

	wantPrefix := []string{
		"./",
		"./bundle.json",
		"./relocation-mapping.json",
		"./artifacts/",
		"./artifacts/layout/",
		"./artifacts/layout/oci-layout",
		"./artifacts/layout/index.json",
	}
	require.Equal(t, wantPrefix, names[:len(wantPrefix)],
		"expected bundle.json, relocation-mapping.json, oci-layout and index.json to precede blobs/")
	for _, name := range names[len(wantPrefix):] {
		require.Contains(t, name, "artifacts/layout/blobs", "expected only blobs/ entries after the metadata files")
	}
}

// TestArchive_CustomTar_NoImages verifies that CustomTar doesn't error when
// there are no images (and so no artifacts/layout/blobs/ directory at all).
func TestArchive_CustomTar_NoImages(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "bundle.json"), []byte(`{}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "relocation-mapping.json"), []byte(`{}`), 0644))

	layoutDir := filepath.Join(dir, "artifacts", "layout")
	require.NoError(t, os.MkdirAll(layoutDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(layoutDir, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(layoutDir, "index.json"), []byte(`{"schemaVersion":2,"manifests":[]}`), 0644))

	ex := &exporter{}
	rc, err := ex.CustomTar(context.Background(), dir, gzip.DefaultCompression)
	require.NoError(t, err)
	defer rc.Close()

	gz, err := gzip.NewReader(rc)
	require.NoError(t, err)
	defer gz.Close()

	var names []string
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		names = append(names, hdr.Name)
	}

	require.Equal(t, []string{
		"./",
		"./bundle.json",
		"./relocation-mapping.json",
		"./artifacts/",
		"./artifacts/layout/",
		"./artifacts/layout/oci-layout",
		"./artifacts/layout/index.json",
	}, names)
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
			err := ex.addImage(context.Background(), baseImage)
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

			err := ex.prepareArtifacts(context.Background(), b)

			require.Equal(t, tc.expectedImgs, collectedImages)
			require.NoError(t, err)
		})
	}

}

type mockCollectingImageStore struct {
	t           *testing.T
	addedImages *[]string
}

func (m mockCollectingImageStore) Add(ctx context.Context, img string) (contentDigest string, err error) {
	*m.addedImages = append(*m.addedImages, img)
	return "digest", nil
}

type mockImageStore struct {
	t        *testing.T
	expected string
}

func (m mockImageStore) Add(ctx context.Context, img string) (contentDigest string, err error) {
	require.Equal(m.t, m.expected, img)
	return "digest", nil
}

// TestArchive_OCILayoutStore_Add verifies that images and image indexes are
// pulled into the archive's OCI layout with their digest preserved, and that
// they can be found again by name, which is how publishing from an archive
// locates them.
func TestArchive_OCILayoutStore_Add(t *testing.T) {
	regSrv := httptest.NewServer(registry.New())
	defer regSrv.Close()
	regHost := strings.TrimPrefix(regSrv.URL, "http://")
	regOpts := cnabtooci.RegistryOptions{InsecureRegistry: true}
	ctx := context.Background()

	img, err := random.Image(1024, 1)
	require.NoError(t, err)
	imgDigest, err := img.Digest()
	require.NoError(t, err)
	imgRef, err := name.ParseReference(fmt.Sprintf("%s/myorg/myapp:v1.0", regHost), regOpts.ToNameOptions()...)
	require.NoError(t, err)
	require.NoError(t, remote.Write(imgRef, img, regOpts.ToRemoteOptions()...))

	idx, err := random.Index(1024, 1, 2)
	require.NoError(t, err)
	idxDigest, err := idx.Digest()
	require.NoError(t, err)
	idxRef, err := name.ParseReference(fmt.Sprintf("%s/myorg/myindex:v1.0", regHost), regOpts.ToNameOptions()...)
	require.NoError(t, err)
	require.NoError(t, remote.WriteIndex(idxRef, idx, regOpts.ToRemoteOptions()...))

	archiveDir := t.TempDir()
	store, err := newOCILayoutStore(archiveDir, cnabtooci.NewRegistry(portercontext.New()), regOpts)
	require.NoError(t, err)

	// Images are referenced by digest in the relocation map
	imgName := fmt.Sprintf("%s/myorg/myapp@%s", regHost, imgDigest)
	gotDigest, err := store.Add(ctx, imgName)
	require.NoError(t, err)
	require.Equal(t, imgDigest.String(), gotDigest)

	idxName := fmt.Sprintf("%s/myorg/myindex@%s", regHost, idxDigest)
	gotDigest, err = store.Add(ctx, idxName)
	require.NoError(t, err)
	require.Equal(t, idxDigest.String(), gotDigest)

	desc, err := findImageInLayout(store.layoutPath, imgName)
	require.NoError(t, err)
	require.Equal(t, imgDigest, desc.Digest)
	require.False(t, desc.MediaType.IsIndex())

	desc, err = findImageInLayout(store.layoutPath, idxName)
	require.NoError(t, err)
	require.Equal(t, idxDigest, desc.Digest)
	require.True(t, desc.MediaType.IsIndex())

	// The image content is in the layout, not just its manifest
	layoutIdx, err := store.layoutPath.ImageIndex()
	require.NoError(t, err)
	layoutImg, err := layoutIdx.Image(imgDigest)
	require.NoError(t, err)
	layers, err := layoutImg.Layers()
	require.NoError(t, err)
	require.Len(t, layers, 1)
	_, err = layers[0].Compressed()
	require.NoError(t, err, "the image layer should be stored in the layout")

	_, err = store.Add(ctx, fmt.Sprintf("%s/myorg/missing:v1.0", regHost))
	require.ErrorAs(t, err, &cnabtooci.ErrNotFound{}, "adding an image that does not exist should fail")
}
