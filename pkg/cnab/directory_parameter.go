package cnab

import (
	"encoding/json"

	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
)

const (
	DirectoryExtensionShortHand    = "directory-parameter"
	DirectoryParameterExtensionKey = PorterExtensionsPrefix + DirectoryExtensionShortHand
)

// DirectoryParameterDefinition represents those parameter options
// That apply exclusively to the directory parameter type
type DirectoryParameterDefinition struct {
	Writeable bool `yaml:"writeable,omitempty"`
	// UID and GID should be ints, however 0 is the default value for int type
	// But is also a realistic value for UID/GID thus we need to make the type interface
	// To detect the case that the values weren't set
	GID interface{} `yaml:"gid,omitempty" json:"gid,omitempty"`
	UID interface{} `yaml:"uid,omitempty" json:"uid,omitempty"`
}

// MountParameterSource represents a parameter using a docker mount
// As a its source with the provided options
type MountParameterSourceDefn struct {
	mount.Mount `yaml:",inline"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
}

// DirectorySources represents the sources available to the directory parameter type
// Currently only mount has been specified, but this could change in the future
type DirectorySources struct {
	Mount MountParameterSourceDefn `yaml:"mount,omitempty" json:"mount,omitempty"`
}
type DirectoryDetails struct {
	DirectorySources
	DirectoryParameterDefinition
	Kind string `json:"kind,omitempty"`
}

// DirectoryParameterExtension indicates that Directory support is required
var DirectoryParameterExtension = RequiredExtension{
	Shorthand: DirectoryExtensionShortHand,
	Key:       DirectoryParameterExtensionKey,
	Reader:    DirectoryParameterReader,
}

// SupportsDirectoryParameters returns true if the bundle supports the
// Directory parameter extension
func (b ExtendedBundle) SupportsDirectoryParameters() bool {
	return b.SupportsExtension(DirectoryParameterExtensionKey)
}

// IsDirType determines if the parameter/credential is of type "directory".
func (b ExtendedBundle) IsDirType(def *definition.Schema) bool {
	return b.SupportsDirectoryParameters() && def.Type == "string" && def.Comment == DirectoryParameterExtensionKey
}

// DirectoryParameterReader is a Reader for the DirectoryParameterExtension.
// The extension maintains the list of directory parameters in the bundle
func DirectoryParameterReader(b ExtendedBundle) (interface{}, error) {
	return b.DirectoryParameterReader()
}

// DirectoryParameterReader is a Reader for the DirectoryParameterExtension.
// This method generates the list of directory parameter names in the bundle.
// The Directory Parameter extension maintains the list of directory parameters in the bundle
func (b ExtendedBundle) DirectoryParameterReader() (interface{}, error) {
	bytes, err := json.Marshal(b.Custom[DirectoryParameterExtensionKey])
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to marshal custom extension %s", DirectoryParameterExtensionKey)
	}
	var dd map[string]DirectoryDetails
	if err = errors.Wrapf(json.Unmarshal(bytes, &dd), "Failed to unmarshal custom extension %s %s", DirectoryParameterExtensionKey, string(bytes)); err != nil {
		return nil, err
	}
	dirs := make([]DirectoryDetails, len(dd))
	i := 0
	for _, dir := range dd {
		dirs[i] = dir
		i++
	}
	return dirs, nil
}

// DirectoryParameterSupport checks if the Directory parameter extension is present
func (e ProcessedExtensions) DirectoryParameterSupport() bool {
	_, extensionRequired := e[DirectoryParameterExtensionKey]
	return extensionRequired
}

// IDToInt converts an interface to an integer. If the id is coercable to an int, returns the value
// Otherwise returns -1
func IDToInt(id interface{}) int {
	if i, ok := id.(int); ok {
		return i
	}

	return -1
}
