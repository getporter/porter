package configadapter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simpleManifestDigest = "4a748b8ac237b4af8f1eb3a327a82dfdd7eb70f3a1126e97a9e5d9b584cd048a"

func TestConfig_GenerateStamp(t *testing.T) {
	// Do not run this test in parallel
	// Still need to figure out what is introducing flakey-ness

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	installedMixins := []mixin.Metadata{
		{Name: "exec", VersionInfo: pkgmgmt.VersionInfo{Version: "v1.2.3"}},
	}

	a := NewManifestConverter(c.Config, m, nil, installedMixins)
	stamp, err := a.GenerateStamp(ctx)
	require.NoError(t, err, "DigestManifest failed")
	assert.Equal(t, simpleManifestDigest, stamp.ManifestDigest)
	assert.Equal(t, map[string]MixinRecord{"exec": {Version: "v1.2.3"}}, stamp.Mixins, "Stamp.Mixins was not populated properly")
	assert.Equal(t, pkg.Version, stamp.Version)
	assert.Equal(t, pkg.Commit, stamp.Commit)

	gotManifestContentsB, err := stamp.DecodeManifest()
	require.NoError(t, err, "DecodeManifest failed")
	wantManifestContentsB, err := c.FileSystem.ReadFile(config.Name)
	require.NoError(t, err, "could not read %s", config.Name)
	assert.Equal(t, string(wantManifestContentsB), string(gotManifestContentsB), "Stamp.EncodedManifest was not popluated and decoded properly")
}

func TestConfig_LoadStamp(t *testing.T) {
	t.Parallel()

	bun := cnab.NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			config.CustomPorterKey: map[string]interface{}{
				"manifestDigest": "somedigest",
				"manifest":       "abc123",
				"mixins": map[string]interface{}{
					"exec": struct{}{},
				},
			},
		},
	})

	stamp, err := LoadStamp(bun)
	require.NoError(t, err)
	assert.Equal(t, "somedigest", stamp.ManifestDigest)
	assert.Equal(t, map[string]MixinRecord{"exec": {}}, stamp.Mixins, "Stamp.Mixins was not populated properly")
	assert.Equal(t, "abc123", stamp.EncodedManifest)
}

func TestConfig_LoadStamp_Invalid(t *testing.T) {
	t.Parallel()

	bun := cnab.NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			config.CustomPorterKey: []string{
				"somedigest",
			},
		},
	})

	stamp, err := LoadStamp(bun)
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
	// Do not run in parallel, it modifies global state
	defer func() { pkg.Version = "" }()

	t.Run("updated version", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

		m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
		require.NoError(t, err, "could not load manifest")

		a := NewManifestConverter(c.Config, m, nil, nil)
		digest, err := a.DigestManifest()
		require.NoError(t, err, "DigestManifest failed")

		pkg.Version = "foo"
		defer func() { pkg.Version = "" }()
		newDigest, err := a.DigestManifest()
		require.NoError(t, err, "DigestManifest failed")
		assert.NotEqual(t, newDigest, digest, "expected the digest to be different due to the updated pkg version")
	})
}

func TestConfig_GenerateStamp_IncludeVersion(t *testing.T) {
	// Do not run this test in parallel
	// Still need to figure out what is introducing flakey-ness

	pkg.Version = "v1.2.3"
	pkg.Commit = "abc123"
	defer func() {
		pkg.Version = ""
		pkg.Commit = ""
	}()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)
	stamp, err := a.GenerateStamp(ctx)
	require.NoError(t, err, "DigestManifest failed")
	assert.Equal(t, "v1.2.3", stamp.Version)
	assert.Equal(t, "abc123", stamp.Commit)
}
