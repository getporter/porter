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
			bundleFileOptions: bundleFileOptions{
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
		name:         "failed to pull image reference",
		opts:         BuildOptions{},
		wantErr:      "failed to pull image",
		wantManifest: "expected-result.yaml",
	},
	}

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockGetCachedImage = mockGetCachedImage

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			manifest := config.Name
			if tc.opts.File != "" {
				manifest = tc.opts.File
			}
			p.TestConfig.TestContext.AddTestFile("testdata/generateManifest/original.yaml", manifest)

			err := tc.opts.Validate(p.Porter)
			require.NoError(t, err)

			if tc.wantErr != "" {
				p.TestRegistry.MockPullImage = mockPullImageFailure
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

func mockPullImageFailure(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) error {
	return fmt.Errorf("failed to pull image %s", ref)
}

func mockGetCachedImage(ctx context.Context, ref cnab.OCIReference) (cnabtooci.ImageSummary, error) {
	sum := types.ImageInspect{
		ID:          "test-id",
		RepoDigests: []string{"test/whalesayd@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f"},
	}
	return cnabtooci.NewImageSummary(ref.String(), sum)
}

func Test_getImageLatestDigest(t *testing.T) {
	defaultMockGetCachedImage := func(ctx context.Context, ref cnab.OCIReference) (cnabtooci.ImageSummary, error) {
		sum := types.ImageInspect{
			ID:          "test-id",
			RepoDigests: []string{"test/repo@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f"},
		}
		return cnabtooci.NewImageSummary(ref.String(), sum)
	}

	testcases := []struct {
		name               string
		imgRef             string
		mockGetCachedImage func(ctx context.Context, ref cnab.OCIReference) (cnabtooci.ImageSummary, error)
		mockPullImage      func(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) error
		wantErr            string
		wantDigest         string
	}{{
		name:       "success",
		imgRef:     "test/repo",
		wantDigest: "sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f",
	}, {
		name:   "non-default image tag",
		imgRef: "test/repo:v0.1.0",
		mockPullImage: func(ctx context.Context, ref cnab.OCIReference, opts cnabtooci.RegistryOptions) error {
			require.True(t, ref.HasTag())
			require.Equal(t, "v0.1.0", ref.Tag())
			return nil
		},
		wantDigest: "sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f",
	}, {
		name:   "failure",
		imgRef: "test/repo",
		mockGetCachedImage: func(ctx context.Context, ref cnab.OCIReference) (cnabtooci.ImageSummary, error) {
			return cnabtooci.ImageSummary{}, errors.New("failed to get cached image")
		},
		wantErr: "failed to get cached image",
	},
	}

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockGetCachedImage = defaultMockGetCachedImage

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ref, err := cnab.ParseOCIReference(tc.imgRef)
			require.NoError(t, err)
			if tc.mockGetCachedImage != nil {
				p.TestRegistry.MockGetCachedImage = tc.mockGetCachedImage
			}
			digest, err := p.getImageLatestDigest(context.Background(), ref)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.Equal(t, tc.wantDigest, digest.String())
		})
	}
}
