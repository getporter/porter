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

func TestInstallation_GetBundleReference(t *testing.T) {
	testcases := []struct {
		name    string
		repo    string
		digest  string
		version string
		wantRef string
		wantErr string
	}{
		{name: "repo missing", wantRef: ""},
		{name: "incomplete reference", repo: "getporter/porter-hello", wantErr: "either BundleDigest or BundleVersion must be specified"},
		{name: "version specified", repo: "getporter/porter-hello", version: "v0.1.1", wantRef: "getporter/porter-hello:v0.1.1"},
		{name: "digest specified", repo: "getporter/porter-hello", digest: "sha256:a881bbc015bade9f11d95a4244888d8e7fa8800f843b43c74cc07c7b7276b062", wantRef: "getporter/porter-hello@sha256:a881bbc015bade9f11d95a4244888d8e7fa8800f843b43c74cc07c7b7276b062"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			i := Installation{
				BundleRepository: tc.repo,
				BundleDigest:     tc.digest,
				BundleVersion:    tc.version,
			}

			ref, ok, err := i.GetBundleReference()
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
