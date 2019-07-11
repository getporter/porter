package configadapter

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simpleManifestDigest = "3f4c199b5b279b5457b00f20b7680e5b1fe644e89005dc8c4f7adf64565252fa"

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
