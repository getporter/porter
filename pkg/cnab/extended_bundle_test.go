package cnab

import (
	"context"
	"testing"

	depsv1ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
	"get.porter.sh/porter/pkg/portercontext"
	porterschema "get.porter.sh/porter/pkg/schema"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtendedBundle_IsPorterBundle(t *testing.T) {
	t.Run("made by porter", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{
			Custom: map[string]interface{}{
				"sh.porter": struct{}{},
			},
		})

		assert.True(t, b.IsPorterBundle())
	})

	t.Run("third party bundle", func(t *testing.T) {
		b := ExtendedBundle{}

		assert.False(t, b.IsPorterBundle())
	})
}

func TestExtendedBundle_IsFileType(t *testing.T) {
	stringDef := &definition.Schema{
		Type: "string",
	}
	fileDef := &definition.Schema{
		Type:            "string",
		ContentEncoding: "base64",
	}
	bun := NewBundle(bundle.Bundle{
		RequiredExtensions: []string{
			FileParameterExtensionKey,
		},
		Definitions: definition.Definitions{
			"string": stringDef,
			"file":   fileDef,
		},
		Parameters: map[string]bundle.Parameter{
			"debug": {
				Definition: "string",
				Required:   true,
			},
			"tfstate": {
				Definition: "file",
			},
		},
	})

	assert.False(t, bun.IsFileType(stringDef), "strings should not be flagged as files")
	assert.True(t, bun.IsFileType(fileDef), "strings+base64 with the file-parameters extension should be categorized as files")

	// Ensure we honor the custom extension
	bun.RequiredExtensions = nil
	assert.False(t, bun.IsFileType(fileDef), "don't categorize as file type when the extension is missing")

	// Ensure we work with old bundles before the extension was created
	bun.Custom = map[string]interface{}{
		"sh.porter": struct{}{},
	}

	assert.True(t, bun.IsFileType(fileDef), "categorize string+base64 in old porter bundles should be categorized as files")
}

func TestExtendedBundle_IsInternalParameter(t *testing.T) {
	bun := NewBundle(bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type: "string",
			},
			"porter-debug": &definition.Schema{
				Type:    "string",
				Comment: PorterInternal,
			},
		},
		Parameters: map[string]bundle.Parameter{
			"foo": {
				Definition: "foo",
			},
			"baz": {
				Definition: "baz",
			},
			"porter-debug": {
				Definition: "porter-debug",
			},
		},
	})

	t.Run("empty bundle", func(t *testing.T) {
		b := ExtendedBundle{}
		require.False(t, b.IsInternalParameter("foo"))
	})

	t.Run("param does not exist", func(t *testing.T) {
		require.False(t, bun.IsInternalParameter("bar"))
	})

	t.Run("definition does not exist", func(t *testing.T) {
		require.False(t, bun.IsInternalParameter("baz"))
	})

	t.Run("is not internal", func(t *testing.T) {
		require.False(t, bun.IsInternalParameter("foo"))
	})

	t.Run("is internal", func(t *testing.T) {
		require.True(t, bun.IsInternalParameter("porter-debug"))
	})
}

func TestExtendedBundle_IsSensitiveParameter(t *testing.T) {
	sensitive := true
	bun := NewBundle(bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type:      "string",
				WriteOnly: &sensitive,
			},
			"porter-debug": &definition.Schema{
				Type:    "string",
				Comment: PorterInternal,
			},
		},
		Parameters: map[string]bundle.Parameter{
			"foo": {
				Definition: "foo",
			},
			"baz": {
				Definition: "baz",
			},
			"porter-debug": {
				Definition: "porter-debug",
			},
		},
	})

	t.Run("empty bundle", func(t *testing.T) {
		b := ExtendedBundle{}
		require.False(t, b.IsSensitiveParameter("foo"))
	})

	t.Run("param does not exist", func(t *testing.T) {
		require.False(t, bun.IsSensitiveParameter("bar"))
	})

	t.Run("definition does not exist", func(t *testing.T) {
		require.False(t, bun.IsSensitiveParameter("baz"))
	})

	t.Run("is not sensitive", func(t *testing.T) {
		require.False(t, bun.IsSensitiveParameter("porter-debug"))
	})

	t.Run("is sensitive", func(t *testing.T) {
		require.True(t, bun.IsSensitiveParameter("foo"))
	})
}

func TestExtendedBundle_GetReferencedRegistries(t *testing.T) {
	t.Run("bundle image in different registry", func(t *testing.T) {
		// Make sure we are looking at the images and the bundle image
		b := NewBundle(bundle.Bundle{
			InvocationImages: []bundle.InvocationImage{
				{BaseImage: bundle.BaseImage{Image: "docker.io/example/mybuns:abc123"}},
			},
			Images: map[string]bundle.Image{
				"nginx": {BaseImage: bundle.BaseImage{Image: "quay.io/library/nginx:latest"}},
				"redis": {BaseImage: bundle.BaseImage{Image: "quay.io/library/redis:latest"}},
				"helm":  {BaseImage: bundle.BaseImage{Image: "ghcr.io/library/helm:latest"}},
			},
		})

		regs, err := b.GetReferencedRegistries()
		require.NoError(t, err, "GetReferencedRegistries failed")
		wantRegs := []string{"docker.io", "ghcr.io", "quay.io"}
		require.Equal(t, wantRegs, regs, "unexpected registries identified in the bundle")
	})

	t.Run("bundle image in same registry", func(t *testing.T) {
		// Make sure that we don't generate duplicate registry entries
		b := NewBundle(bundle.Bundle{
			InvocationImages: []bundle.InvocationImage{
				{BaseImage: bundle.BaseImage{Image: "ghcr.io/example/mybuns:abc123"}},
			},
			Images: map[string]bundle.Image{
				"nginx": {BaseImage: bundle.BaseImage{Image: "quay.io/library/nginx:latest"}},
				"redis": {BaseImage: bundle.BaseImage{Image: "quay.io/library/redis:latest"}},
				"helm":  {BaseImage: bundle.BaseImage{Image: "ghcr.io/library/helm:latest"}},
			},
		})

		regs, err := b.GetReferencedRegistries()
		require.NoError(t, err, "GetReferencedRegistries failed")
		wantRegs := []string{"ghcr.io", "quay.io"}
		require.Equal(t, wantRegs, regs, "unexpected registries identified in the bundle")
	})
}

func TestValidate(t *testing.T) {
	testcases := []struct {
		name       string
		version    string
		strategy   porterschema.CheckStrategy
		hasWarning bool
		wantErr    string
	}{
		{name: "older version", strategy: porterschema.CheckStrategyExact, version: "1.0.0"},
		{name: "current version", strategy: porterschema.CheckStrategyExact, version: "1.2.0"},
		{name: "unsupported version", strategy: porterschema.CheckStrategyExact, version: "1.3.0", wantErr: "invalid"},
		{name: "custom version check strategy", strategy: porterschema.CheckStrategyMajor, version: "1.1.1", hasWarning: true, wantErr: "WARNING"},
	}

	cxt := portercontext.NewTestContext(t)
	defer cxt.Close()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBundle(bundle.Bundle{
				SchemaVersion: SchemaVersion(tc.version),
				InvocationImages: []bundle.InvocationImage{
					{BaseImage: bundle.BaseImage{}},
				},
			})

			err := b.Validate(cxt.Context, tc.strategy)
			if tc.wantErr != "" && !tc.hasWarning {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Contains(t, cxt.GetError(), tc.wantErr)
		})
	}
}

func TestExtendedBundle_ResolveDependencies(t *testing.T) {
	t.Parallel()

	bun := NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			DependenciesV1ExtensionKey: depsv1ext.Dependencies{
				Requires: map[string]depsv1ext.Dependency{
					"mysql": {
						Bundle: "getporter/mysql:5.7",
					},
					"nginx": {
						Bundle: "localhost:5000/nginx:1.19",
					},
				},
			},
		},
	})

	eb := ExtendedBundle{}
	locks, err := eb.ResolveDependencies(context.Background(), bun)
	require.NoError(t, err)
	require.Len(t, locks, 2)

	var mysql DependencyLock
	var nginx DependencyLock
	for _, lock := range locks {
		switch lock.Alias {
		case "mysql":
			mysql = lock
		case "nginx":
			nginx = lock
		}
	}

	assert.Equal(t, "getporter/mysql:5.7", mysql.Reference)
	assert.Equal(t, "localhost:5000/nginx:1.19", nginx.Reference)
}

// mockRegistryProvider mocks the registry for testing
type mockRegistryProvider struct {
	tags map[string][]string // map repository to tags
	err  error
}

func (m *mockRegistryProvider) ListTags(ctx context.Context, ref OCIReference, opts interface{}) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	repo := ref.Repository()
	if tags, ok := m.tags[repo]; ok {
		return tags, nil
	}
	return []string{}, nil
}

func TestExtendedBundle_ResolveVersion(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name        string
		dep         depsv1ext.Dependency
		wantVersion string
		wantError   string
	}{
		{name: "pinned version",
			dep:         depsv1ext.Dependency{Bundle: "mysql:5.7"},
			wantVersion: "5.7"},
		{name: "unimplemented range",
			dep:       depsv1ext.Dependency{Bundle: "mysql", Version: &depsv1ext.DependencyVersion{Ranges: []string{"1 - 1.5"}}},
			wantError: "not implemented"},
		{name: "default tag to latest",
			dep:         depsv1ext.Dependency{Bundle: "getporterci/porter-test-only-latest"},
			wantVersion: "latest"},
		{name: "no default tag",
			dep:       depsv1ext.Dependency{Bundle: "getporterci/porter-test-no-default-tag"},
			wantError: "no tag was specified"},
		{name: "default tag to highest semver",
			dep:         depsv1ext.Dependency{Bundle: "getporterci/porter-test-with-versions", Version: &depsv1ext.DependencyVersion{Ranges: nil, AllowPrereleases: true}},
			wantVersion: "v1.3-beta1"},
		{name: "default tag to highest semver, explicitly excluding prereleases",
			dep:         depsv1ext.Dependency{Bundle: "getporterci/porter-test-with-versions", Version: &depsv1ext.DependencyVersion{Ranges: nil, AllowPrereleases: false}},
			wantVersion: "v1.2"},
		{name: "default tag to highest semver, excluding prereleases by default",
			dep:         depsv1ext.Dependency{Bundle: "getporterci/porter-test-with-versions"},
			wantVersion: "v1.2"},
	}

	// Create mock registry with test data
	mockReg := &mockRegistryProvider{
		tags: map[string][]string{
			"getporterci/porter-test-only-latest":      {"latest"},
			"getporterci/porter-test-no-default-tag":   {"not-a-semver"},
			"getporterci/porter-test-with-versions":    {"latest", "v1.0", "v1.1", "v1.2", "v1.3-beta1"},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eb := ExtendedBundle{}.WithRegistry(mockReg, nil)
			version, err := eb.ResolveVersion(context.Background(), "mysql", tc.dep)
			if tc.wantError != "" {
				require.Error(t, err, "ResolveVersion should have returned an error")
				assert.Contains(t, err.Error(), tc.wantError)
			} else {
				require.NoError(t, err, "ResolveVersion should not have returned an error")

				assert.Equal(t, tc.wantVersion, version.Tag(), "incorrect version resolved")
			}
		})
	}
}
