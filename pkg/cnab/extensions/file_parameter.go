package extensions

import (
	"github.com/cnabio/cnab-go/bundle"
	cnabbundle "github.com/cnabio/cnab-go/bundle"
)

const (
	// FileParameterExtensionShortHand is the short suffix of the FileParameterExtensionKey.
	FileParameterExtensionShortHand = "file-parameters"

	// FileParameterExtensionKey represents the full key for the File Parameter extension.
	FileParameterExtensionKey = PorterExtensionsPrefix + FileParameterExtensionShortHand
)

// DockerExtension represents a required extension enabling access to the host Docker daemon
var FileParameterExtension = RequiredExtension{
	Shorthand: FileParameterExtensionShortHand,
	Key:       FileParameterExtensionKey,
	Reader:    FileParameterReader,
}

// FileParameterReader is a Reader for the FileParameterExtension.
// The extension does not have any data, it's presence indicates that
// parameters of type "file" should be supported by the tooling.
func FileParameterReader(_ cnabbundle.Bundle) (interface{}, error) {
	return nil, nil
}

// SupportsFileParameters checks if the bundle supports file parameters.
func SupportsFileParameters(b bundle.Bundle) bool {
	if SupportsExtension(b, FileParameterExtensionKey) {
		return true
	}

	// Porter has always supported this but didn't have the extension declared
	// TODO(v1): Remove this logic in v1.0?
	return IsPorterBundle(b)
}

// FileParameterSupport checks if the docker extension is present and returns its
// extension configuration.
func (e ProcessedExtensions) FileParameterSupport() bool {
	_, extensionRequired := e[FileParameterExtensionKey]
	return extensionRequired
}
