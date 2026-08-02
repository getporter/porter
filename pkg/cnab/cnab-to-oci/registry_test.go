package cnabtooci

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
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
			sum, err := NewImageSummaryFromInspect(cnab.MustParseOCIReference(tt.imgRef), client.ImageInspectResult{InspectResponse: tt.imageSummary})
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

func TestRegistry_GetRemoteImageDigest(t *testing.T) {
	regSrv := httptest.NewServer(registry.New())
	defer regSrv.Close()
	regHost := strings.TrimPrefix(regSrv.URL, "http://")

	regOpts := RegistryOptions{InsecureRegistry: true}
	r := NewRegistry(portercontext.New())
	ctx := context.Background()

	t.Run("image exists", func(t *testing.T) {
		pushRef, err := name.ParseReference(regHost+"/myorg/myapp:v1.0", regOpts.ToNameOptions()...)
		require.NoError(t, err)

		img, err := random.Image(1024, 1)
		require.NoError(t, err)
		require.NoError(t, remote.Write(pushRef, img, regOpts.ToRemoteOptions()...))

		wantDigest, err := img.Digest()
		require.NoError(t, err)

		ref := cnab.MustParseOCIReference(regHost + "/myorg/myapp:v1.0")
		gotDigest, err := r.GetRemoteImageDigest(ctx, ref, regOpts)
		require.NoError(t, err)
		assert.Equal(t, wantDigest.String(), gotDigest.String())
	})

	t.Run("image does not exist", func(t *testing.T) {
		ref := cnab.MustParseOCIReference(regHost + "/myorg/missing:v1.0")
		_, err := r.GetRemoteImageDigest(ctx, ref, regOpts)
		require.ErrorIs(t, err, ErrNotFound{})
	})
}

// rejectTagDeleteHandler wraps a registry handler and rejects DELETE
// requests targeting a tag (as opposed to a digest), mimicking the behavior
// of the reference distribution/distribution registry implementation, which
// most self-hosted registries run and which does not support the OCI
// distribution spec's DELETE-by-tag endpoint.
type rejectTagDeleteHandler struct {
	inner http.Handler
}

func (h *rejectTagDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		if _, tag, ok := strings.Cut(r.URL.Path, "/manifests/"); ok && !strings.HasPrefix(tag, "sha256:") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"errors":[{"code":"DIGEST_INVALID","message":"provided digest did not match uploaded content"}]}`))
			return
		}
	}
	h.inner.ServeHTTP(w, r)
}

func TestRegistry_DeleteImageTag(t *testing.T) {
	regOpts := RegistryOptions{InsecureRegistry: true}
	r := NewRegistry(portercontext.New())
	ctx := context.Background()

	t.Run("registry supports deleting a tag directly", func(t *testing.T) {
		regSrv := httptest.NewServer(registry.New())
		defer regSrv.Close()
		regHost := strings.TrimPrefix(regSrv.URL, "http://")

		pushRef, err := name.ParseReference(regHost+"/myorg/myapp:v1.0", regOpts.ToNameOptions()...)
		require.NoError(t, err)
		img, err := random.Image(1024, 1)
		require.NoError(t, err)
		require.NoError(t, remote.Write(pushRef, img, regOpts.ToRemoteOptions()...))

		ref := cnab.MustParseOCIReference(regHost + "/myorg/myapp:v1.0")
		require.NoError(t, r.DeleteImageTag(ctx, ref, regOpts))

		_, err = remote.Head(pushRef, regOpts.ToRemoteOptions()...)
		require.Error(t, err, "expected the tag to no longer resolve")
	})

	t.Run("falls back to a placeholder manifest when the registry rejects deleting a tag directly", func(t *testing.T) {
		regSrv := httptest.NewServer(&rejectTagDeleteHandler{inner: registry.New()})
		defer regSrv.Close()
		regHost := strings.TrimPrefix(regSrv.URL, "http://")

		pushRef, err := name.ParseReference(regHost+"/myorg/myapp:v1.0", regOpts.ToNameOptions()...)
		require.NoError(t, err)
		img, err := random.Image(1024, 1)
		require.NoError(t, err)
		require.NoError(t, remote.Write(pushRef, img, regOpts.ToRemoteOptions()...))
		wantDigest, err := img.Digest()
		require.NoError(t, err)

		ref := cnab.MustParseOCIReference(regHost + "/myorg/myapp:v1.0")
		require.NoError(t, r.DeleteImageTag(ctx, ref, regOpts))

		// The in-memory fake registry used here doesn't cascade-delete a tag
		// when its manifest is deleted by digest (unlike real registries,
		// which is exactly why this fallback works against them -- see the
		// "falls back..." case). So the strongest thing this fake lets us
		// assert is that the tag no longer points at the original content:
		// it was overwritten with the throwaway placeholder as part of the
		// fallback, and a plain DELETE-by-tag was rejected, so nothing else
		// could have removed the original digest from under it.
		desc, err := remote.Head(pushRef, regOpts.ToRemoteOptions()...)
		require.NoError(t, err)
		assert.NotEqual(t, wantDigest.String(), desc.Digest.String(), "the tag should no longer point at the original image")

		// The original manifest, referenced by digest, must remain untouched --
		// it may still be referenced elsewhere, e.g. by a bundle we just published.
		digestRef, err := name.ParseReference(regHost+"/myorg/myapp@"+wantDigest.String(), regOpts.ToNameOptions()...)
		require.NoError(t, err)
		desc, err = remote.Head(digestRef, regOpts.ToRemoteOptions()...)
		require.NoError(t, err, "the original manifest should still be reachable by digest")
		assert.Equal(t, wantDigest.String(), desc.Digest.String())
	})

	t.Run("rejects a digest-based reference", func(t *testing.T) {
		ref := cnab.MustParseOCIReference("example.com/myorg/myapp@sha256:6b5a28ccbb76f12ce771a23757880c6083234255c5ba191fca1c5db1f71c1687")
		err := r.DeleteImageTag(ctx, ref, regOpts)
		require.Error(t, err, "DeleteImageTag should refuse a digest-based reference")
	})

	t.Run("does not overwrite the tag when the registry rejects delete for a reason other than DIGEST_INVALID", func(t *testing.T) {
		regSrv := httptest.NewServer(&rejectAllDeleteHandler{inner: registry.New(), statusCode: http.StatusForbidden})
		defer regSrv.Close()
		regHost := strings.TrimPrefix(regSrv.URL, "http://")

		pushRef, err := name.ParseReference(regHost+"/myorg/myapp:v1.0", regOpts.ToNameOptions()...)
		require.NoError(t, err)
		img, err := random.Image(1024, 1)
		require.NoError(t, err)
		require.NoError(t, remote.Write(pushRef, img, regOpts.ToRemoteOptions()...))
		wantDigest, err := img.Digest()
		require.NoError(t, err)

		ref := cnab.MustParseOCIReference(regHost + "/myorg/myapp:v1.0")
		err = r.DeleteImageTag(ctx, ref, regOpts)
		require.Error(t, err, "expected DeleteImageTag to surface the original error instead of falling back")

		// The tag must still point at the original content: since the failure
		// wasn't the known DIGEST_INVALID case, DeleteImageTag must not have
		// attempted the overwrite workaround.
		desc, err := remote.Head(pushRef, regOpts.ToRemoteOptions()...)
		require.NoError(t, err)
		assert.Equal(t, wantDigest.String(), desc.Digest.String(), "the tag should be untouched when the fallback isn't attempted")
	})
}

// rejectAllDeleteHandler wraps a registry handler and rejects every DELETE
// request with the given status code, mimicking a registry that denies
// manifest deletion outright (e.g. due to permissions or the feature being
// disabled), as opposed to the specific DIGEST_INVALID response used by
// rejectTagDeleteHandler to mimic unsupported DELETE-by-tag.
type rejectAllDeleteHandler struct {
	inner      http.Handler
	statusCode int
}

func (h *rejectAllDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		w.WriteHeader(h.statusCode)
		_, _ = w.Write([]byte(`{"errors":[{"code":"DENIED","message":"access denied"}]}`))
		return
	}
	h.inner.ServeHTTP(w, r)
}
