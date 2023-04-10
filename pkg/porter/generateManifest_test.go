package porter

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/test"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/require"
)

func Test_generateInternalManifest(t *testing.T) {
	testcases := []struct {
		name         string
		opts         BuildOptions
		wantErr      string
		wantManifest string
	}{{
		name:         "no opts",
		opts:         BuildOptions{},
		wantManifest: "expected-result.yaml",
	}, {
		name: "--file set",
		opts: BuildOptions{
			BundleDefinitionOptions: BundleDefinitionOptions{
				File: "alternate.yaml",
			},
		},
		wantManifest: "expected-result.yaml",
	}, {
		name:         "name set",
		opts:         BuildOptions{metadataOpts: metadataOpts{Name: "newname"}},
		wantManifest: "new-name.yaml",
	}, {
		name:         "version set",
		opts:         BuildOptions{metadataOpts: metadataOpts{Version: "1.0.0"}},
		wantManifest: "new-version.yaml",
	}, {
		name:         "name and value set",
		opts:         BuildOptions{metadataOpts: metadataOpts{Name: "newname", Version: "1.0.0"}},
		wantManifest: "all-fields.yaml",
	}, {
		name:         "custom input set",
		opts:         BuildOptions{Customs: []string{"key1=editedValue1", "key2.nestedKey2=editedValue2"}},
		wantManifest: "custom-input.yaml",
	}, {
		name:         "failed to fetch image metadata",
		opts:         BuildOptions{},
		wantErr:      "failed to fetch image metadata",
		wantManifest: "expected-result.yaml",
	},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			manifest := config.Name
			if tc.opts.File != "" {
				manifest = tc.opts.File
			}
			p.TestConfig.TestContext.AddTestFile("testdata/generateManifest/original.yaml", manifest)

			err := tc.opts.Validate(p.Porter)
			require.NoError(t, err)

			if tc.wantErr == "" {
				p.TestRegistry.MockGetCachedImage = mockGetCachedImage
			} else {
				p.TestRegistry.MockGetImageMetadata = mockGetImageMetadataFailure
			}

			err = p.generateInternalManifest(context.Background(), tc.opts)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			goldenFile := filepath.Join("testdata/generateManifest", tc.wantManifest)
			p.TestConfig.TestContext.AddTestFile(goldenFile, tc.wantManifest)
			got, err := p.FileSystem.ReadFile(build.LOCAL_MANIFEST)
			require.NoError(t, err)
			test.CompareGoldenFile(t, goldenFile, string(got))
		})
	}
}

func mockGetImageMetadataFailure(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) (cnabtooci.ImageMetadata, error) {
	return cnabtooci.ImageMetadata{}, fmt.Errorf("failed to fetch image metadata %s", ref)
}

func mockGetCachedImage(ctx context.Context, ref cnab.OCIReference) (cnabtooci.ImageMetadata, error) {
	sum := types.ImageInspect{
		ID:          "test-id",
		RepoDigests: []string{"test/whalesayd@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f"},
	}
	return cnabtooci.NewImageSummaryFromInspect(ref, sum)
}

func Test_getImageLatestDigest(t *testing.T) {
	defaultMockGetCachedImage := func(ctx context.Context, ref cnab.OCIReference) (cnabtooci.ImageMetadata, error) {
		sum := types.ImageInspect{
			ID:          "test-id",
			RepoDigests: []string{"test/repo@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f"},
		}
		return cnabtooci.NewImageSummaryFromInspect(ref, sum)
	}

	testcases := []struct {
		name                 string
		imgRef               string
		mockGetCachedImage   func(ctx context.Context, ref cnab.OCIReference) (cnabtooci.ImageMetadata, error)
		mockGetImageMetadata func(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) (cnabtooci.ImageMetadata, error)
		wantErr              string
		wantDigest           string
	}{{
		name:       "success",
		imgRef:     "test/repo",
		wantDigest: "sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f",
	}, {
		name:   "non-default image tag",
		imgRef: "test/repo:v0.1.0",
		mockGetCachedImage: func(ctx context.Context, ref cnab.OCIReference) (cnabtooci.ImageMetadata, error) {
			return cnabtooci.ImageMetadata{}, fmt.Errorf("not in cache")
		},
		mockGetImageMetadata: func(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) (cnabtooci.ImageMetadata, error) {
			require.True(t, ref.HasTag())
			require.Equal(t, "v0.1.0", ref.Tag())
			return cnabtooci.ImageMetadata{
				Reference: ref,
				RepoDigests: []string{
					"test/repo@sha256:d2134b0be91e9e293bc872b719538440e5f933e9828cd96430c85d904afb5aa9",
				},
			}, nil
		},
		wantDigest: "sha256:d2134b0be91e9e293bc872b719538440e5f933e9828cd96430c85d904afb5aa9",
	}, {
		name:   "failure",
		imgRef: "test/repo",
		mockGetCachedImage: func(ctx context.Context, ref cnab.OCIReference) (cnabtooci.ImageMetadata, error) {
			return cnabtooci.ImageMetadata{}, errors.New("failed to get cached image")
		},
		mockGetImageMetadata: mockGetImageMetadataFailure,
		wantErr:              "failed to fetch image metadata",
	},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			ref, err := cnab.ParseOCIReference(tc.imgRef)
			require.NoError(t, err)

			if tc.mockGetCachedImage != nil {
				p.TestRegistry.MockGetCachedImage = tc.mockGetCachedImage
			} else {
				p.TestRegistry.MockGetCachedImage = defaultMockGetCachedImage
			}
			if tc.mockGetImageMetadata != nil {
				p.TestRegistry.MockGetImageMetadata = tc.mockGetImageMetadata
			}

			regOpts := cnabtooci.RegistryOptions{}
			digest, err := p.getImageDigest(context.Background(), ref, regOpts)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.Equal(t, tc.wantDigest, digest.String())
		})
	}
}
