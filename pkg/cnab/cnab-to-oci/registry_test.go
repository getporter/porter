package cnabtooci_test

import (
	"testing"

	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"github.com/docker/docker/api/types"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestImageSummary(t *testing.T) {
	type expectedOutput struct {
		imageRef string
		digest   string
	}

	testcases := []struct {
		name         string
		imgRef       string
		imageSummary types.ImageSummary
		expected     expectedOutput
		expectedErr  string
	}{
		{
			name:         "successful initialization",
			imgRef:       "test/image:latest",
			imageSummary: types.ImageSummary{ID: "test", RepoDigests: []string{"test/image@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"}},
			expected:     expectedOutput{imageRef: "test/image:latest", digest: "sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"},
			expectedErr:  "",
		},
		{
			name:         "invalid image reference",
			imgRef:       "test-",
			imageSummary: types.ImageSummary{ID: "test", RepoDigests: []string{"test/image@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"}},
			expectedErr:  "invalid reference format",
		},
		{
			name:         "empty repo digests",
			imgRef:       "test/image:latest",
			imageSummary: types.ImageSummary{ID: "test", RepoDigests: []string{}},
			expectedErr:  "failed to get digest",
			expected: expectedOutput{
				imageRef: "test/image:latest",
			},
		},
		{
			name:         "failed to find valid digest",
			imgRef:       "test/image:latest",
			imageSummary: types.ImageSummary{ID: "test", RepoDigests: []string{"test/image-another-repo@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"}},
			expectedErr:  digest.ErrDigestInvalidFormat.Error(),
			expected: expectedOutput{
				imageRef: "test/image:latest",
			},
		},
	}

	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sum, err := cnabtooci.NewImageSummary(tt.imgRef, tt.imageSummary)
			if tt.expected.imageRef == "" {
				require.ErrorContains(t, err, tt.expectedErr)
				return
			}
			require.Equal(t, sum.GetImageReference().String(), tt.expected.imageRef)
			digest, err := sum.Digest()
			if tt.expected.digest == "" {
				require.ErrorContains(t, err, tt.expectedErr)
				return
			}
			require.Equal(t, tt.expected.digest, digest.String())
		})
	}
}
