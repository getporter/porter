package cnab

import (
	"encoding/json"

	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
)

const (
	// DockerExtensionShortHand is the short suffix of the DockerExtensionKey.
	DockerExtensionShortHand = "docker"

	// DockerExtensionKey represents the full key for the Docker Extension
	DockerExtensionKey = OfficialExtensionsPrefix + DockerExtensionShortHand

	// DockerExtensionSchema represents the schema for the Docker Extension
	DockerExtensionSchema = "schema/io-cnab-docker.schema.json"
)

// DockerExtension represents a required extension enabling access to the host Docker daemon
var DockerExtension = RequiredExtension{
	Shorthand: DockerExtensionShortHand,
	Key:       DockerExtensionKey,
	Schema:    "schema/io-cnab-docker.schema.json",
	Reader:    DockerExtensionReader,
}

// Docker describes the set of custom extension metadata associated with the Docker extension
type Docker struct {
	// Privileged represents whether or not the Docker container should run as --privileged
	Privileged bool `json:"privileged,omitempty"`
	Mounts []mount.Mount `json:"mounts,omitempty"`
	Network string `json:"network,omitempty"`
	CapAdd []string `json:"capadd,omitempty"`
	CapDrop []string `json:"capdrop,omitempty"`
}

// DockerExtensionReader is a Reader for the DockerExtension,
// which reads from the applicable section in the provided bundle and
// returns the raw data in the form of an interface
func DockerExtensionReader(bun ExtendedBundle) (interface{}, error) {
	return bun.DockerExtensionReader()
}

// DockerExtensionReader is a Reader for the DockerExtension,
// which reads from the applicable section in the provided bundle and
// returns the raw data in the form of an interface
func (b ExtendedBundle) DockerExtensionReader() (interface{}, error) {
	data, ok := b.Custom[DockerExtensionKey]
	if !ok {
		return nil, errors.New("no custom extension configuration found")
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal the untyped %q extension data %q",
			DockerExtensionKey, string(dataB))
	}

	dha := Docker{}
	err = json.Unmarshal(dataB, &dha)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal the %q extension %q",
			DockerExtensionKey, string(dataB))
	}

	return dha, nil
}

// GetDocker checks if the docker extension is present and returns its
// extension configuration.
func (e ProcessedExtensions) GetDocker() (dockerExt Docker, dockerRequired bool, err error) {
	ext, extensionRequired := e[DockerExtensionKey]

	dockerExt, ok := ext.(Docker)
	if !ok && extensionRequired {
		return Docker{}, extensionRequired, errors.Errorf("unable to parse Docker extension config: %+v", dockerExt)
	}

	return dockerExt, extensionRequired, nil
}

// SupportsDocker checks if the bundle supports docker.
func (b ExtendedBundle) SupportsDocker() bool {
	return b.SupportsExtension(DockerExtensionKey)
}
