package config

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simpleManifestDigest = "b3adf8240f3289c8e7a6076ce7538bc7381d5c1d581f993b627f13d9102a3617"

func TestConfig_ComputeManifestDigest(t *testing.T) {
	c := NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	stamp := c.GenerateStamp(c.Manifest)
	assert.Equal(t, simpleManifestDigest, stamp.ManifestDigest)
}

func TestConfig_LoadStamp(t *testing.T) {
	c := NewTestConfig(t)

	bun := &bundle.Bundle{
		Custom: map[string]interface{}{
			CustomBundleKey: map[string]interface{}{
				"manifestDigest": simpleManifestDigest,
			},
		},
	}

	stamp, err := c.LoadStamp(bun)
	require.NoError(t, err)
	assert.Equal(t, simpleManifestDigest, stamp.ManifestDigest)
}

func TestConfig_LoadStamp_Invalid(t *testing.T) {
	c := NewTestConfig(t)

	bun := &bundle.Bundle{
		Custom: map[string]interface{}{
			CustomBundleKey: []string{
				simpleManifestDigest,
			},
		},
	}

	stamp, err := c.LoadStamp(bun)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not unmarshal the porter stamp")
	assert.Nil(t, stamp)
}
