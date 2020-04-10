package configadapter

import (
	"testing"

	"get.porter.sh/porter/pkg/manifest"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simpleManifestDigest = "6154be0570d30a2b654d02258d9fa3004a72df9d5b70dafc0efce73c6818dc8f"

func TestConfig_GenerateStamp(t *testing.T) {
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
