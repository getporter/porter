package porter

import (
	"github.com/deislabs/cnab-go/bundle"
)

// Registry handles talking with an OCI registry.
type Registry interface {
	// PullBundle pulls a bundle from an OCI registry.
	PullBundle(tag string, insecureRegistry bool) (*bundle.Bundle, error)

	// PushBundle pushes a bundle to an OCI registry.
	PushBundle(bun *bundle.Bundle, tag string, insecureRegistry bool) error

	// PushInvocationImage pushes the invocation image from the Docker image cache to the specified location
	// the expected format of the invocationImage is REGISTRY/NAME:TAG.
	// Returns the image digest from the registry.
	PushInvocationImage(invocationImage string) (string, error)

	// Copy copies the original image given by origImg to the new image provided by newImg,
	// returning the new digest and/or an error
	Copy(origImg, newImg string) (string, error)
}
