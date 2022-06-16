package cnab

import (
	"testing"

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
