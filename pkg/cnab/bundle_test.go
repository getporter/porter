package cnab

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBundleReference(t *testing.T) {
	testcases := []struct {
		Name       string
		Reference  string
		WantRepo   string
		WantTag    string
		WantDigest string
		WantErr    string
	}{
		{Name: "version", Reference: "getporter/porter-hello:v0.1.0", WantRepo: "getporter/porter-hello", WantTag: "v0.1.0"},
		{Name: "digest", Reference: "getporter/porter-hello@sha256:88d68ef0bdb9cedc6da3a8e341a33e5d2f8bb19d0cf7ec3f1060d3f9eb73cae9", WantRepo: "getporter/porter-hello", WantDigest: "sha256:88d68ef0bdb9cedc6da3a8e341a33e5d2f8bb19d0cf7ec3f1060d3f9eb73cae9"},
		{Name: "invalid", Reference: "@v1", WantErr: "invalid bundle reference"},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			gotRepo, gotTag, gotDigest, err := ParseBundleReference(tc.Reference)
			if tc.WantErr != "" {
				require.Error(t, err, "expected ParseBundleReference to fail")
				assert.Contains(t, err.Error(), tc.WantErr)
			} else {
				require.NoError(t, err, "expected ParseBundleReference to succeed")
				assert.Equal(t, tc.WantRepo, gotRepo, "incorrect repo")
				assert.Equal(t, tc.WantTag, gotTag, "incorrect tag")
				assert.Equal(t, tc.WantDigest, gotDigest, "incorrect digest")
			}
		})
	}
}
