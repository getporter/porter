package cnab

const (
	// FileParameterExtensionShortHand is the short suffix of the FileParameterExtensionKey.
	FileParameterExtensionShortHand = "file-parameters"

	// FileParameterExtensionKey represents the full key for the File Parameter extension.
	FileParameterExtensionKey = PorterExtensionsPrefix + FileParameterExtensionShortHand
)

// FileParameterExtension represents a required extension that indicates that the bundle
// requires support for parameters of type "file"
var FileParameterExtension = RequiredExtension{
	Shorthand: FileParameterExtensionShortHand,
	Key:       FileParameterExtensionKey,
	Reader:    FileParameterReader,
}

// FileParameterReader is a Reader for the FileParameterExtension.
// The extension does not have any data, its presence indicates that
// parameters of type "file" should be supported by the tooling.
func FileParameterReader(b ExtendedBundle) (interface{}, error) {
	return b.FileParameterReader()
}

// FileParameterReader is a Reader for the FileParameterExtension.
// The extension does not have any data, its presence indicates that
// parameters of type "file" should be supported by the tooling.
func (b ExtendedBundle) FileParameterReader() (interface{}, error) {
	return nil, nil
}

// SupportsFileParameters checks if the bundle supports file parameters.
func (b ExtendedBundle) SupportsFileParameters() bool {
	if b.SupportsExtension(FileParameterExtensionKey) {
		return true
	}

	// Porter has always supported this but didn't have the extension declared.
	return b.IsPorterBundle()
}

// FileParameterSupport checks if the file parameter extension
// is present.
func (e ProcessedExtensions) FileParameterSupport() bool {
	_, extensionRequired := e[FileParameterExtensionKey]
	return extensionRequired
}
