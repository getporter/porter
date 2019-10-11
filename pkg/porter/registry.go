package porter

import (
	"github.com/deislabs/cnab-go/bundle"
)

// Registry handles talking with an OCI registry.
type Registry interface {
	// PullBundle pulls a bundle from an OCI registry and returns the relocation map
	PullBundle(tag string, insecureRegistry bool) (*bundle.Bundle, map[string]string, error)

	// PushBundle pushes a bundle to an OCI registry and returns the relocation map
	PushBundle(bun *bundle.Bundle, tag string, insecureRegistry bool) (map[string]string, error)

	// PushInvocationImage pushes the invocation image from the Docker image cache to the specified location
	// the expected format of the invocationImage is REGISTRY/NAME:TAG.
	// Returns the image digest from the registry.
	PushInvocationImage(invocationImage string) (string, error)
}
