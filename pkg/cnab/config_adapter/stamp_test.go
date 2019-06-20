package configadapter

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simpleManifestDigest = "2960616546bba09016d6849e56134ec07bd873597326ad0093811b0992c51769"

func TestConfig_ComputeManifestDigest(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("../../config/testdata/simple.porter.yaml", config.Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	a := ManifestConverter{
		Context:  c.Context,
		Manifest: c.Manifest,
	}
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
