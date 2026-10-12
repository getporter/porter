package cnab

import (
	"encoding/json"
	"fmt"
	"slices"
)

const (
	// DirectoryParameterExtensionShortHand is the short suffix of the DirectoryParameterExtensionKey.
	DirectoryParameterExtensionShortHand = "directory-parameters"

	// DirectoryParameterExtensionKey represents the full key for the Directory Parameter extension.
	DirectoryParameterExtensionKey = "sh.porter." + DirectoryParameterExtensionShortHand
)

// DirectoryParameterExtension represents a required extension that indicates that the bundle
// requires support for parameters and outputs of type "directory".
var DirectoryParameterExtension = RequiredExtension{
	Shorthand: DirectoryParameterExtensionShortHand,
	Key:       DirectoryParameterExtensionKey,
	Reader:    DirectoryParameterReader,
}

// DirectoryParameters lists the parameters and outputs of the bundle that
// are directories. Directories are stored in the bundle as a base64 encoded
// tar.gz archive, just like files, so the definition alone cannot tell them apart.
type DirectoryParameters struct {
	// Parameters is the list of parameter names that are directories.
	Parameters []string `json:"parameters,omitempty" mapstructure:"parameters"`

	// Outputs is the list of output names that are directories.
	Outputs []string `json:"outputs,omitempty" mapstructure:"outputs"`
}

// DirectoryParameterReader is a Reader for the DirectoryParameterExtension.
func DirectoryParameterReader(b ExtendedBundle) (interface{}, error) {
	return b.DirectoryParameterReader()
}

// DirectoryParameterReader is a Reader for the DirectoryParameterExtension.
func (b ExtendedBundle) DirectoryParameterReader() (interface{}, error) {
	raw, ok := b.Custom[DirectoryParameterExtensionKey]
	if !ok {
		return DirectoryParameters{}, nil
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the untyped %q extension, %v: %w", DirectoryParameterExtensionKey, raw, err)
	}

	var dirs DirectoryParameters
	if err := json.Unmarshal(data, &dirs); err != nil {
		return nil, fmt.Errorf("could not unmarshal the %q extension %q: %w", DirectoryParameterExtensionKey, string(data), err)
	}
	return dirs, nil
}

// SupportsDirectoryParameters checks if the bundle declares the directory parameter extension.
func (b ExtendedBundle) SupportsDirectoryParameters() bool {
	return b.SupportsExtension(DirectoryParameterExtensionKey)
}

// directoryParameters returns the directory parameters declared by the bundle,
// or an empty set when the bundle doesn't use the extension.
func (b ExtendedBundle) directoryParameters() DirectoryParameters {
	if !b.SupportsDirectoryParameters() {
		return DirectoryParameters{}
	}
	dirs, err := b.DirectoryParameterReader()
	if err != nil {
		return DirectoryParameters{}
	}
	return dirs.(DirectoryParameters)
}

// IsDirectoryParameter determines if the named parameter is of type "directory".
func (b ExtendedBundle) IsDirectoryParameter(name string) bool {
	return slices.Contains(b.directoryParameters().Parameters, name)
}

// IsDirectoryOutput determines if the named output is of type "directory".
func (b ExtendedBundle) IsDirectoryOutput(name string) bool {
	return slices.Contains(b.directoryParameters().Outputs, name)
}
