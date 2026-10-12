package cnab

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
)

func TestDirectoryTypeByName_SameNameParamAndOutput(t *testing.T) {
	b64 := &definition.Schema{Type: "string", ContentEncoding: "base64"}
	// "data" is a directory parameter, and a file output with the same name
	b := NewBundle(bundle.Bundle{
		RequiredExtensions: []string{FileParameterExtensionKey, DirectoryParameterExtensionKey},
		Custom: map[string]interface{}{
			DirectoryParameterExtensionKey: map[string]interface{}{"parameters": []string{"data"}},
		},
	})

	assert.Equal(t, "directory", b.GetParameterTypeByName("data", b64))
	assert.Equal(t, "file", b.GetOutputTypeByName("data", b64))
	assert.Equal(t, "file", b.GetParameterTypeByName("other", b64))
}

func TestDirectoryParameters_NotDeclared(t *testing.T) {
	b := NewBundle(bundle.Bundle{})
	assert.False(t, b.IsDirectoryParameter("data"))
	assert.False(t, b.IsDirectoryOutput("data"))
}
