package extensions

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
)

func TestIsPorterBundle(t *testing.T) {
	t.Run("made by porter", func(t *testing.T) {
		b := bundle.Bundle{
			Custom: map[string]interface{}{
				"sh.porter": struct{}{},
			},
		}

		assert.True(t, IsPorterBundle(b))
	})

	t.Run("third party bundle", func(t *testing.T) {
		b := bundle.Bundle{}

		assert.False(t, IsPorterBundle(b))
	})
}

func TestIsFileType(t *testing.T) {
	stringDef := &definition.Schema{
		Type: "string",
	}
	fileDef := &definition.Schema{
		Type:            "string",
		ContentEncoding: "base64",
	}
	bun := bundle.Bundle{
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
	}

	assert.False(t, IsFileType(bun, stringDef), "strings should not be flagged as files")
	assert.True(t, IsFileType(bun, fileDef), "strings+base64 with the file-parameters extension should be categorized as files")

	// Ensure we honor the custom extension
	bun.RequiredExtensions = nil
	assert.False(t, IsFileType(bun, fileDef), "don't categorize as file type when the extension is missing")

	// Ensure we work with old bundles before the extension was created
	bun.Custom = map[string]interface{}{
		"sh.porter": struct{}{},
	}

	assert.True(t, IsFileType(bun, fileDef), "categorize string+base64 in old porter bundles should be categorized as files")
}
