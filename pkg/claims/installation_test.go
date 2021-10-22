package claims

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallation_String(t *testing.T) {
	i := Installation{Name: "mybun"}
	assert.Equal(t, "/mybun", i.String())

	i.Namespace = "dev"
	assert.Equal(t, "dev/mybun", i.String())
}

func TestOCIReferenceParts_GetBundleReference(t *testing.T) {
	testcases := []struct {
		name    string
		repo    string
		digest  string
		version string
		tag     string
		wantRef string
		wantErr string
	}{
		{name: "repo missing", wantRef: ""},
		{name: "incomplete reference", repo: "getporter/porter-hello", wantErr: "Invalid bundle reference"},
		{name: "version specified", repo: "getporter/porter-hello", version: "v0.1.1", wantRef: "getporter/porter-hello:v0.1.1"},
		{name: "digest specified", repo: "getporter/porter-hello", digest: "sha256:a881bbc015bade9f11d95a4244888d8e7fa8800f843b43c74cc07c7b7276b062", wantRef: "getporter/porter-hello@sha256:a881bbc015bade9f11d95a4244888d8e7fa8800f843b43c74cc07c7b7276b062"},
		{name: "tag specified", repo: "getporter/porter-hello", tag: "latest", wantRef: "getporter/porter-hello:latest"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b := OCIReferenceParts{
				Repository: tc.repo,
				Digest:     tc.digest,
				Version:    tc.version,
				Tag:        tc.tag,
			}

			ref, ok, err := b.GetBundleReference()
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			} else if tc.wantRef != "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantRef, ref.String())
			} else {
				require.NoError(t, err)
				require.False(t, ok)
			}
		})
	}
}
