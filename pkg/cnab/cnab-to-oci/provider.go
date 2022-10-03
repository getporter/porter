package cnabtooci

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/google/go-containerregistry/pkg/crane"
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
	// Use ErrNotFound to detect if the failure is because the image is not in the local Docker cache.
	GetCachedImage(ctx context.Context, ref cnab.OCIReference) (ImageSummary, error)

	// ListTags returns all tags defined on the specified repository.
	ListTags(ctx context.Context, repo cnab.OCIReference, opts RegistryOptions) ([]string, error)

	// PullImage pulls an image from an OCI registry and returns the image's digest
	PullImage(ctx context.Context, image cnab.OCIReference, opts RegistryOptions) error

	// GetBundleMetadata returns information about a bundle in a registry
	// Use ErrNotFound to detect if the error is because the bundle is not in the registry.
	GetBundleMetadata(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (BundleMetadata, error)
}

// RegistryOptions is the set of options for interacting with an OCI registry.
type RegistryOptions struct {
	// InsecureRegistry allows connecting to an unsecured registry or one without verifiable certificates.
	InsecureRegistry bool
}

func (o RegistryOptions) toCraneOptions() []crane.Option {
	var result []crane.Option
	if o.InsecureRegistry {
		transport := GetInsecureRegistryTransport()
		result = []crane.Option{crane.Insecure, crane.WithTransport(transport)}
	}
	return result
}
