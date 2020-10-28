package cnabtooci

import (
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
)

// Registry handles talking with an OCI registry.
type RegistryProvider interface {
	// PullBundle pulls a bundle from an OCI registry.
	PullBundle(tag string, insecureRegistry bool) (bundle.Bundle, *relocation.ImageRelocationMap, error)

	// PushBundle pushes a bundle to an OCI registry.
	PushBundle(bun bundle.Bundle, tag string, insecureRegistry bool) (*relocation.ImageRelocationMap, error)

	// PushInvocationImage pushes the invocation image from the Docker image cache to the specified location
	// the expected format of the invocationImage is REGISTRY/NAME:TAG.
	// Returns the image digest from the registry.
	PushInvocationImage(invocationImage string) (string, error)
}
