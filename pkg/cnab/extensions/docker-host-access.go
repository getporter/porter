package extensions

import (
	"encoding/json"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

const (
	// DockerHostAccessKey represents the full key for the DockerHostAccess Extension
	DockerHostAccessKey = "io.cnab.docker-host-access"
	// DockerHostAccessSchema represents the schema for the DockerHostAccess Extension
	DockerHostAccessSchema = "TODO"
)

// DockerHostAccessExtension represents a required extension enabling access to the host Docker daemon
var DockerHostAccessExtension = RequiredExtension{
	Shorthand: "docker-host-access",
	Key:       DockerHostAccessKey,
	Schema:    DockerHostAccessSchema,
	Reader:    DockerHostAccessReader,
}

// DockerHostAccess describes the set of custom extension metadata associated with the DockerHostAccess extension
type DockerHostAccess struct {
	// Privileged represents whether or not the Docker container should run as --privileged
	Privileged bool `json:"privileged,omitempty"`
}

// DockerHostAccessReader is a Reader for the DockerHostAccessExtension,
// which reads from the applicable section in the provided bundle and
// returns a the raw data in the form of an interface
func DockerHostAccessReader(bun *bundle.Bundle) (interface{}, error) {
	data, ok := bun.Custom[DockerHostAccessKey]
	if !ok {
		// TODO: we should error out here, right?
		return nil, nil
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal the untyped %q extension data %q",
			DockerHostAccessKey, string(dataB))
	}

	dha := &DockerHostAccess{}
	err = json.Unmarshal(dataB, dha)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal the %q extension %q",
			DockerHostAccessKey, string(dataB))
	}

	return dha, nil
}
