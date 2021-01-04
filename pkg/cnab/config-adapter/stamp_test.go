package configadapter

import (
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/manifest"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simpleManifestDigest = "b3be65771034c64a0d49d2c8a4ac3103a1ec12d6e41015ef57861fd913f72ecf"

func TestConfig_GenerateStamp(t *testing.T) {
	// Do not run in parallel, it's flakey and I haven't figured out why yet
	// I assume cruft in bin and the timing of when the test is run

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("../../manifest/testdata/simple.porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Context, m, nil, nil)
	stamp, err := a.GenerateStamp()
	require.NoError(t, err, "DigestManifest failed")
	assert.Equal(t, simpleManifestDigest, stamp.ManifestDigest)
	assert.Equal(t, map[string]MixinRecord{"exec": {}}, stamp.Mixins, "Stamp.Mixins was not populated properly")

	gotManifestContentsB, err := stamp.DecodeManifest()
	require.NoError(t, err, "DecodeManifest failed")
	wantManifestContentsB, err := c.FileSystem.ReadFile(config.Name)
	require.NoError(t, err, "could not read %s", config.Name)
	assert.Equal(t, string(wantManifestContentsB), string(gotManifestContentsB), "Stamp.EncodedManifest was not popluated and decoded properly")
}

func TestConfig_LoadStamp(t *testing.T) {
	t.Parallel()

	bun := &bundle.Bundle{
		Custom: map[string]interface{}{
			config.CustomPorterKey: map[string]interface{}{
				"manifestDigest": simpleManifestDigest,
				"manifest":       "abc123",
				"mixins": map[string]interface{}{
					"exec": struct{}{},
				},
			},
		},
	}

	stamp, err := LoadStamp(*bun)
	require.NoError(t, err)
	assert.Equal(t, simpleManifestDigest, stamp.ManifestDigest)
	assert.Equal(t, map[string]MixinRecord{"exec": {}}, stamp.Mixins, "Stamp.Mixins was not populated properly")
	assert.Equal(t, "abc123", stamp.EncodedManifest)
}

func TestConfig_LoadStamp_Invalid(t *testing.T) {
	t.Parallel()

	bun := &bundle.Bundle{
		Custom: map[string]interface{}{
			config.CustomPorterKey: []string{
				simpleManifestDigest,
			},
		},
	}

	stamp, err := LoadStamp(*bun)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not unmarshal the porter stamp")
	assert.Equal(t, Stamp{}, stamp)
}

func TestStamp_DecodeManifest(t *testing.T) {
	t.Parallel()

	t.Run("manifest populated", func(t *testing.T) {
		t.Parallel()

		c := config.NewTestConfig(t)
		s := Stamp{
			EncodedManifest: "bmFtZTogaGVsbG8=", // name: hello
		}

		data, err := s.DecodeManifest()
		require.NoError(t, err, "DecodeManifest failed")

		m, err := manifest.UnmarshalManifest(c.TestContext.Context, data)
		require.NoError(t, err, "UnmarshalManifest failed")

		require.NotNil(t, m, "expected manifest to be populated")
		assert.Equal(t, "hello", m.Name, "expected the manifest name to be populated")
	})

	t.Run("manifest empty", func(t *testing.T) {
		t.Parallel()

		s := Stamp{}

		data, err := s.DecodeManifest()
		require.EqualError(t, err, "no Porter manifest was embedded in the bundle")

		assert.Nil(t, data, "No manifest data should be returned")
	})

	t.Run("manifest invalid", func(t *testing.T) {
		t.Parallel()

		s := Stamp{
			EncodedManifest: "name: hello", // this should be base64 encoded
		}

		data, err := s.DecodeManifest()
		require.Error(t, err, "DecodeManifest should fail for invalid data")

		assert.Contains(t, err.Error(), "could not base64 decode the manifest in the stamp")
		assert.Nil(t, data, "No manifest data should be returned")
	})

}

func TestConfig_DigestManifest(t *testing.T) {
	t.Parallel()

	t.Run("updated invocation image", func(t *testing.T) {
		t.Parallel()

		c := config.NewTestConfig(t)
		c.TestContext.AddTestFile("../../manifest/testdata/simple.porter.yaml", config.Name)

		m, err := manifest.LoadManifestFrom(c.Context, config.Name)
		require.NoError(t, err, "could not load manifest")

		a := NewManifestConverter(c.Context, m, nil, nil)
		digest, err := a.DigestManifest()
		require.NoError(t, err, "DigestManifest failed")

		m.Image = "newpublishregistry/porter-hello:v0.1.0"
		newDigest, err := a.DigestManifest()
		require.NoError(t, err, "DigestManifest failed")
		assert.NotEqual(t, newDigest, digest, "expected the digest to be different due to the updated image")
	})

	t.Run("updated version", func(t *testing.T) {
		t.Parallel()

		c := config.NewTestConfig(t)
		c.TestContext.AddTestFile("../../manifest/testdata/simple.porter.yaml", config.Name)

		m, err := manifest.LoadManifestFrom(c.Context, config.Name)
		require.NoError(t, err, "could not load manifest")

		a := NewManifestConverter(c.Context, m, nil, nil)
		digest, err := a.DigestManifest()
		require.NoError(t, err, "DigestManifest failed")

		pkg.Version = "foo"
		newDigest, err := a.DigestManifest()
		require.NoError(t, err, "DigestManifest failed")
		assert.NotEqual(t, newDigest, digest, "expected the digest to be different due to the updated pkg version")
	})
}
