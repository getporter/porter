package cnabtooci

import (
	"net/http"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/docker/docker/api/types/image"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
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
		imageSummary image.InspectResponse
		expected     expectedOutput
		expectedErr  string
	}{
		{
			name:         "successful initialization",
			imgRef:       "test/image:latest",
			imageSummary: image.InspectResponse{ID: "test", RepoDigests: []string{"test/image@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"}},
			expected:     expectedOutput{imageRef: "test/image:latest", digest: "sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"},
			expectedErr:  "",
		},
		{
			name:         "empty repo digests",
			imgRef:       "test/image:latest",
			imageSummary: image.InspectResponse{ID: "test", RepoDigests: []string{}},
			expectedErr:  "failed to get digest",
			expected: expectedOutput{
				imageRef: "test/image:latest",
			},
		},
		{
			name:         "failed to find valid digest",
			imgRef:       "test/image:latest",
			imageSummary: image.InspectResponse{ID: "test", RepoDigests: []string{"test/image-another-repo@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687"}},
			expectedErr:  "cannot find image digest for desired repo",
			expected: expectedOutput{
				imageRef: "test/image:latest",
			},
		},
	}

	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sum, err := NewImageSummaryFromInspect(cnab.MustParseOCIReference(tt.imgRef), tt.imageSummary)
			if tt.expected.hasInitErr {
				require.ErrorContains(t, err, tt.expectedErr)
				return
			}
			require.Equal(t, sum.Reference.String(), tt.expected.imageRef)
			digest, err := sum.GetRepositoryDigest()
			if tt.expected.digest == "" {
				require.ErrorContains(t, err, tt.expectedErr)
				return
			}
			require.Equal(t, tt.expected.digest, digest.String())
		})
	}
}

func TestNewImageSummaryFromDescriptor(t *testing.T) {
	ref := cnab.MustParseOCIReference("localhost:5000/whalesayd:latest")
	origRepoDigest := digest.Digest("sha256:499f71eec2e3bd78f26c268bbf5b2a65f73b96216fac4a89b86b5ebf115527b6")

	s, err := NewImageSummaryFromDigest(ref, origRepoDigest)
	require.NoError(t, err, "NewImageSummaryFromDigest failed")

	// Locate the repository digest associated with the reference and validate that it matches what we input
	repoDigest, err := s.GetRepositoryDigest()
	require.NoError(t, err, "failed to get repository digest for image summary")
	assert.Equal(t, origRepoDigest.String(), repoDigest.String())
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
