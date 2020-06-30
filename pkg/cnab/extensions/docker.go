package extensions

import (
	"encoding/json"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

const (
	// DockerExtensionKey represents the full key for the Docker Extension
	DockerExtensionKey = "io.cnab.docker"
	// DockerExtensionSchema represents the schema for the Docker Extension
	DockerExtensionSchema = "schema/io-cnab-docker.schema.json"
)

// DockerExtension represents a required extension enabling access to the host Docker daemon
var DockerExtension = RequiredExtension{
	Shorthand: "docker",
	Key:       DockerExtensionKey,
	Schema:    DockerExtensionSchema,
	Reader:    DockerExtensionReader,
}

// Docker describes the set of custom extension metadata associated with the Docker extension
type Docker struct {
	// Privileged represents whether or not the Docker container should run as --privileged
	Privileged bool `json:"privileged,omitempty"`

	// UseHostNetwork is set to true if the Docker container should use the host network
	UseHostNetwork bool `json:"useHostNetwork,omitempty"`
}

// DockerExtensionReader is a Reader for the DockerExtension,
// which reads from the applicable section in the provided bundle and
// returns a the raw data in the form of an interface
func DockerExtensionReader(bun *bundle.Bundle) (interface{}, error) {
	data, ok := bun.Custom[DockerExtensionKey]
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

// GetDockerExtension checks if the docker extension is present and returns its
// extension configuration.
func (e ProcessedExtensions) GetDockerExtension() (dockerExt Docker, dockerRequired bool, err error) {
	ext, extensionRequired := e[DockerExtensionKey]

	dockerExt, ok := ext.(Docker)
	if !ok && extensionRequired {
		return Docker{}, extensionRequired, errors.Errorf("unable to parse Docker extension config: %+v", dockerExt)
	}

	return dockerExt, extensionRequired, nil
}
