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

func TestDirectoryParameterReader_MissingConfiguration(t *testing.T) {
	b := NewBundle(bundle.Bundle{RequiredExtensions: []string{DirectoryParameterExtensionKey}})
	_, err := b.DirectoryParameterReader()
	assert.ErrorContains(t, err, "no custom extension configuration found")
}

func TestDirectoryParameters_ImplyFileParameters(t *testing.T) {
	// A bundle that only requires the directory extension still gets file handling
	b := NewBundle(bundle.Bundle{
		RequiredExtensions: []string{DirectoryParameterExtensionKey},
		Custom: map[string]interface{}{
			DirectoryParameterExtensionKey: map[string]interface{}{"parameters": []string{"data"}},
		},
	})
	assert.True(t, b.SupportsFileParameters())
	assert.True(t, b.IsFileType(&definition.Schema{Type: "string", ContentEncoding: "base64"}))
}
