package cnabtooci

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/opencontainers/go-digest"
)

// RegistryProvider handles talking with an OCI registry.
type RegistryProvider interface {
	// PullBundle pulls a bundle from an OCI registry.
	PullBundle(ctx context.Context, ref cnab.OCIReference, insecureRegistry bool) (cnab.BundleReference, error)

	// PushBundle pushes a bundle to an OCI registry.
	PushBundle(ctx context.Context, bundleRef cnab.BundleReference, insecureRegistry bool) (cnab.BundleReference, error)

	// PushInvocationImage pushes the invocation image from the Docker image cache to the specified location
	// the expected format of the invocationImage is REGISTRY/NAME:TAG.
	// Returns the image digest from the registry.
	PushInvocationImage(ctx context.Context, invocationImage string) (digest.Digest, error)

	// IsImageCached checks whether a particular invocation image exists in the local image cache.
	IsImageCached(ctx context.Context, invocationImage string) (bool, error)

	// ListTags returns all tags defined on the specified repository.
	ListTags(ctx context.Context, repository string) ([]string, error)

	// PullImage pulls a image from an OCI registry and returns the image's digest
	PullImage(ctx context.Context, image string) (string, error)
}
