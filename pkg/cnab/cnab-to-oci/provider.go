package cnabtooci

import (
	"get.porter.sh/porter/pkg/cnab"
	"github.com/opencontainers/go-digest"
)

// Registry handles talking with an OCI registry.
type RegistryProvider interface {
	// PullBundle pulls a bundle from an OCI registry.
	PullBundle(ref cnab.OCIReference, insecureRegistry bool) (cnab.BundleReference, error)

	// PushBundle pushes a bundle to an OCI registry.
	PushBundle(bundleRef cnab.BundleReference, insecureRegistry bool) (cnab.BundleReference, error)

	// PushInvocationImage pushes the invocation image from the Docker image cache to the specified location
	// the expected format of the invocationImage is REGISTRY/NAME:TAG.
	// Returns the image digest from the registry.
	PushInvocationImage(invocationImage string) (digest.Digest, error)

	IsInvocationImageExists(invocationImage string) (bool, error)
}
