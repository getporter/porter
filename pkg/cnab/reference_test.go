package cnab

import (
	"encoding/json"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOCIReference(t *testing.T) {
	tests := []struct {
		Name             string
		Ref              OCIReference
		ExpectedTag      string
		ExpectedDigest   string
		ExpectedVersion  string
		ExpectedRepo     string
		ExpectedRegistry string
	}{
		{
			Name:             "valid digested reference",
			Ref:              MustParseOCIReference("jeremyrickard/porter-do-bundle@sha256:a808aa4e3508d7129742eefda938249574447cce5403dc12d4cbbfe7f4f31e58"),
			ExpectedRegistry: "docker.io",
			ExpectedRepo:     "jeremyrickard/porter-do-bundle",
			ExpectedDigest:   "sha256:a808aa4e3508d7129742eefda938249574447cce5403dc12d4cbbfe7f4f31e58",
		},
		{
			Name:             "tagged reference",
			Ref:              MustParseOCIReference("jeremyrickard/porter-do-bundle:v0.1.0"),
			ExpectedRegistry: "docker.io",
			ExpectedRepo:     "jeremyrickard/porter-do-bundle",
			ExpectedTag:      "v0.1.0",
			ExpectedVersion:  "0.1.0",
		},
		{
			Name:             "no tag",
			Ref:              MustParseOCIReference("ghcr.io/jeremyrickard/porter-do-bundle"),
			ExpectedRepo:     "ghcr.io/jeremyrickard/porter-do-bundle",
			ExpectedRegistry: "ghcr.io",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(t, test.ExpectedRepo, test.Ref.Repository())

			assert.Equal(t, test.ExpectedRegistry, test.Ref.Registry())

			assert.Equal(t, test.ExpectedDigest, test.Ref.Digest().String())
			assert.Equal(t, test.ExpectedDigest != "", test.Ref.HasDigest())

			assert.Equal(t, test.ExpectedTag, test.Ref.Tag())
			assert.Equal(t, test.ExpectedTag != "", test.Ref.HasTag())

			assert.Equal(t, test.ExpectedVersion, test.Ref.Version())
			assert.Equal(t, test.ExpectedVersion != "", test.Ref.HasVersion())
		})
	}
}

func TestParseOCIReference(t *testing.T) {
	testcases := []struct {
		Name       string
		Reference  string
		WantRepo   string
		WantTag    string
		WantDigest digest.Digest
		WantErr    string
	}{
		{Name: "version", Reference: "getporter/porter-hello:v0.1.0", WantRepo: "getporter/porter-hello", WantTag: "v0.1.0"},
		{Name: "digest", Reference: "getporter/porter-hello@sha256:88d68ef0bdb9cedc6da3a8e341a33e5d2f8bb19d0cf7ec3f1060d3f9eb73cae9", WantRepo: "getporter/porter-hello", WantDigest: "sha256:88d68ef0bdb9cedc6da3a8e341a33e5d2f8bb19d0cf7ec3f1060d3f9eb73cae9"},
		{Name: "invalid", Reference: "@v1", WantErr: "invalid reference format"},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			ref, err := ParseOCIReference(tc.Reference)
			if tc.WantErr != "" {
				require.Error(t, err, "expected ParseOCIReference to fail")
				assert.Contains(t, err.Error(), tc.WantErr)
			} else {
				require.NoError(t, err, "expected ParseOCIReference to succeed")
				assert.Equal(t, tc.WantRepo, ref.Repository(), "incorrect repo")
				assert.Equal(t, tc.WantTag, ref.Tag(), "incorrect tag")
				assert.Equal(t, tc.WantDigest, ref.Digest(), "incorrect digest")
			}
		})
	}
}

func TestOCIReference_MarshalJSON(t *testing.T) {
	r := MustParseOCIReference("getporter/porter-hello:v0.1.1")
	data, err := json.Marshal(r)
	require.NoError(t, err)
	assert.Equal(t, `"getporter/porter-hello:v0.1.1"`, string(data))
}

func TestOCIReference_UnmarshalJSON(t *testing.T) {
	ref := `"getporter/porter-hello:v0.1.1"`
	var r OCIReference
	err := json.Unmarshal([]byte(ref), &r)
	require.NoError(t, err)
	assert.Equal(t, "getporter/porter-hello:v0.1.1", r.String())
}

func TestOCIReference_WithVersion(t *testing.T) {
	t.Run("prefixed semver", func(t *testing.T) {
		ref := MustParseOCIReference("getporter/porter-hello")

		result, err := ref.WithVersion("v1.2.3")
		require.NoError(t, err)
		assert.Equal(t, "v1.2.3", result.Tag())
	})

	t.Run("unprefixed semver", func(t *testing.T) {
		ref := MustParseOCIReference("getporter/porter-hello")

		result, err := ref.WithVersion("1.2.3")
		require.NoError(t, err)
		assert.Equal(t, "v1.2.3", result.Tag())
	})

	t.Run("invalid semver", func(t *testing.T) {
		ref := MustParseOCIReference("getporter/porter-hello")

		_, err := ref.WithVersion("oops")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid bundle version")
	})
}

func TestOCIReference_WithTag(t *testing.T) {
	ref := MustParseOCIReference("getporter/porter-hello")

	result, err := ref.WithTag("latest")
	require.NoError(t, err)
	assert.Equal(t, "latest", result.Tag())
}

func TestOCIReference_WithDigest(t *testing.T) {
	t.Run("valid digest", func(t *testing.T) {
		ref := MustParseOCIReference("getporter/porter-hello")

		result, err := ref.WithDigest("sha256:a808aa4e3508d7129742eefda938249574447cce5403dc12d4cbbfe7f4f31e58")
		require.NoError(t, err)
		assert.Equal(t, "sha256:a808aa4e3508d7129742eefda938249574447cce5403dc12d4cbbfe7f4f31e58", result.Digest().String())
	})

	t.Run("invalid digest", func(t *testing.T) {
		ref := MustParseOCIReference("getporter/porter-hello")

		_, err := ref.WithDigest("oops")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid digest")
	})
}
