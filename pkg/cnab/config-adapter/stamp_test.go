package configadapter

import (
	"testing"

	"get.porter.sh/porter/pkg/manifest"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simpleManifestDigest = "74ef9c4abc7dfc41566ba28d9fcfab9c70baa0ca5ca434c71298689487f554e8"

func TestConfig_ComputeManifestDigest(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("../../manifest/testdata/simple.porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Context, m, nil, nil)
	stamp := a.GenerateStamp()
	assert.Equal(t, simpleManifestDigest, stamp.ManifestDigest)
}

func TestConfig_LoadStamp(t *testing.T) {
	bun := &bundle.Bundle{
		Custom: map[string]interface{}{
			config.CustomBundleKey: map[string]interface{}{
				"manifestDigest": simpleManifestDigest,
			},
		},
	}

	stamp, err := LoadStamp(bun)
	require.NoError(t, err)
	assert.Equal(t, simpleManifestDigest, stamp.ManifestDigest)
}

func TestConfig_LoadStamp_Invalid(t *testing.T) {
	bun := &bundle.Bundle{
		Custom: map[string]interface{}{
			config.CustomBundleKey: []string{
				simpleManifestDigest,
			},
		},
	}

	stamp, err := LoadStamp(bun)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not unmarshal the porter stamp")
	assert.Nil(t, stamp)
}
