package cnabtooci

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/opencontainers/go-digest"
)

// RegistryProvider handles talking with an OCI registry.
type RegistryProvider interface {
	// PullBundle pulls a bundle from an OCI registry.
	PullBundle(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (cnab.BundleReference, error)

	// PushBundle pushes a bundle to an OCI registry.
	PushBundle(ctx context.Context, ref cnab.BundleReference, opts RegistryOptions) (cnab.BundleReference, error)

	// PushImage pushes the image from the Docker image cache to the specified location
	// the expected format of the image is REGISTRY/NAME:TAG.
	// Returns the image digest from the registry.
	PushImage(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (digest.Digest, error)

	// GetCachedImage returns a particular image from the local image cache.
	GetCachedImage(ctx context.Context, ref cnab.OCIReference) (ImageSummary, error)

	// ListTags returns all tags defined on the specified repository.
	ListTags(ctx context.Context, repo cnab.OCIReference, opts RegistryOptions) ([]string, error)

	// PullImage pulls a image from an OCI registry and returns the image's digest
	PullImage(ctx context.Context, image cnab.OCIReference, opts RegistryOptions) error
}

// RegistryOptions is the set of options for interacting with an OCI registry.
type RegistryOptions struct {
	// InsecureRegistry allows connecting to an unsecured registry or one without verifiable certificates.
	InsecureRegistry bool
}
