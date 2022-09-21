package cnabtooci

import (
	"net/http"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/docker/docker/api/types"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/stretchr/testify/require"
)

func TestImageSummary(t *testing.T) {
	type expectedOutput struct {
		imageRef   string
		digest     string
		hasInitErr bool
	}

	testcases := []struct {
		name         string
		imgRef       string
		imageSummary types.ImageInspect
		expected     expectedOutput
		expectedErr  string
	}{
		{
			name:         "successful initialization",
			imgRef:       "test/image:latest",
			imageSummary: types.ImageInspect{ID: "test", RepoDigests: []string{"test/image@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"}},
			expected:     expectedOutput{imageRef: "test/image:latest", digest: "sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"},
			expectedErr:  "",
		},
		{
			name:         "invalid image reference",
			imgRef:       "test-",
			imageSummary: types.ImageInspect{ID: "test", RepoDigests: []string{"test/image@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"}},
			expected:     expectedOutput{hasInitErr: true},
			expectedErr:  "invalid reference format",
		},
		{
			name:         "empty repo digests",
			imgRef:       "test/image:latest",
			imageSummary: types.ImageInspect{ID: "test", RepoDigests: []string{}},
			expectedErr:  "failed to get digest",
			expected: expectedOutput{
				imageRef: "test/image:latest",
			},
		},
		{
			name:         "failed to find valid digest",
			imgRef:       "test/image:latest",
			imageSummary: types.ImageInspect{ID: "test", RepoDigests: []string{"test/image-another-repo@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"}},
			expectedErr:  "cannot find image digest for desired repo",
			expected: expectedOutput{
				imageRef: "test/image:latest",
			},
		},
	}

	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sum, err := NewImageSummary(tt.imgRef, tt.imageSummary)
			if tt.expected.hasInitErr {
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

func TestAsNotFoundError(t *testing.T) {
	ref := cnab.MustParseOCIReference("example.com/mybuns:v1.2.3")
	t.Run("404", func(t *testing.T) {
		srcErr := &transport.Error{
			StatusCode: http.StatusNotFound,
		}
		result := asNotFoundError(srcErr, ref)
		require.NotNil(t, result)
		require.Equal(t, ErrNotFound{Reference: ref}, result)
	})
	t.Run("401", func(t *testing.T) {
		srcErr := &transport.Error{
			StatusCode: http.StatusUnauthorized,
		}
		result := asNotFoundError(srcErr, ref)
		require.Nil(t, result)
	})
}
