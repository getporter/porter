package cnab

import (
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	porterschema "get.porter.sh/porter/pkg/schema"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/schema"
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
	t.Run("invocation image in different registry", func(t *testing.T) {
		// Make sure we are looking at the images and the invocation image
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

	t.Run("invocation image in same registry", func(t *testing.T) {
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
				SchemaVersion: schema.Version(tc.version),
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
