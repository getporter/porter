package configadapter

import (
	"context"
	"sort"
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
	testcases := []struct {
		name         string
		preserveTags bool
	}{
		{name: "not preserving tags", preserveTags: false},
		{name: "preserving tags", preserveTags: true},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := config.NewTestConfig(t)
			c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

			ctx := context.Background()
			m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
			require.NoError(t, err, "could not load manifest")

			installedMixins := []mixin.Metadata{
				{Name: "exec", VersionInfo: pkgmgmt.VersionInfo{Version: "v1.2.3"}},
			}

			a := NewManifestConverter(c.Config, m, nil, installedMixins, tc.preserveTags)
			stamp, err := a.GenerateStamp(ctx, tc.preserveTags)
			require.NoError(t, err, "DigestManifest failed")
			assert.Equal(t, simpleManifestDigest, stamp.ManifestDigest)
			assert.Equal(t, map[string]MixinRecord{"exec": {Name: "exec", Version: "v1.2.3"}}, stamp.Mixins, "Stamp.Mixins was not populated properly")
			assert.Equal(t, pkg.Version, stamp.Version)
			assert.Equal(t, pkg.Commit, stamp.Commit)
			assert.Equal(t, tc.preserveTags, stamp.PreserveTags)

			gotManifestContentsB, err := stamp.DecodeManifest()
			require.NoError(t, err, "DecodeManifest failed")
			wantManifestContentsB, err := c.FileSystem.ReadFile(config.Name)
			require.NoError(t, err, "could not read %s", config.Name)
			assert.Equal(t, string(wantManifestContentsB), string(gotManifestContentsB), "Stamp.EncodedManifest was not popluated and decoded properly")
		})
	}
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
				"preserveTags": true,
			},
		},
	})

	stamp, err := LoadStamp(bun)
	require.NoError(t, err)
	assert.Equal(t, "somedigest", stamp.ManifestDigest)
	assert.Equal(t, map[string]MixinRecord{"exec": {}}, stamp.Mixins, "Stamp.Mixins was not populated properly")
	assert.Equal(t, "abc123", stamp.EncodedManifest)
	assert.Equal(t, true, stamp.PreserveTags)
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

		a := NewManifestConverter(c.Config, m, nil, nil, false)
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

	a := NewManifestConverter(c.Config, m, nil, nil, false)
	stamp, err := a.GenerateStamp(ctx, false)
	require.NoError(t, err, "DigestManifest failed")
	assert.Equal(t, "v1.2.3", stamp.Version)
	assert.Equal(t, "abc123", stamp.Commit)
}

func TestMixinRecord_Sort(t *testing.T) {
	records := MixinRecords{
		{Name: "helm", Version: "0.1.13"},
		{Name: "helm", Version: "v0.1.2"},
		{Name: "testmixin", Version: "1.2.3"},
		{Name: "exec", Version: "2.1.0"},
		// These won't parse as valid semver, so just sort them by the string representation instead
		{
			Name:    "az",
			Version: "invalid-version2",
		},
		{
			Name:    "az",
			Version: "invalid-version1",
		},
	}

	sort.Sort(records)

	wantRecords := MixinRecords{
		{
			Name:    "az",
			Version: "invalid-version1",
		},
		{
			Name:    "az",
			Version: "invalid-version2",
		},
		{
			Name:    "exec",
			Version: "2.1.0",
		},
		{
			Name:    "helm",
			Version: "v0.1.2",
		},
		{
			Name:    "helm",
			Version: "0.1.13",
		},
		{
			Name:    "testmixin",
			Version: "1.2.3",
		},
	}

	assert.Equal(t, wantRecords, records)
}
