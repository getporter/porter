package cnabtooci

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/docker/docker/api/types"
	"github.com/opencontainers/go-digest"
)

var _ RegistryProvider = &TestRegistry{}

type TestRegistry struct {
	MockPullBundle        func(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (cnab.BundleReference, error)
	MockPushBundle        func(ctx context.Context, ref cnab.BundleReference, opts RegistryOptions) (bundleReference cnab.BundleReference, err error)
	MockPushImage         func(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (imageDigest digest.Digest, err error)
	MockGetCachedImage    func(ctx context.Context, ref cnab.OCIReference) (ImageMetadata, error)
	MockListTags          func(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) ([]string, error)
	MockPullImage         func(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) error
	MockGetBundleMetadata func(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (BundleMetadata, error)
	MockGetImageMetadata  func(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (ImageMetadata, error)
	cache                 map[string]ImageMetadata
}

func NewTestRegistry() *TestRegistry {
	return &TestRegistry{
		cache: make(map[string]ImageMetadata),
	}
}

func (t TestRegistry) PullBundle(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (cnab.BundleReference, error) {
	if t.MockPullBundle != nil {
		return t.MockPullBundle(ctx, ref, opts)
	}
	sum, _ := NewImageSummaryFromInspect(ref, types.ImageInspect{ID: cnab.NewULID()})
	t.cache[ref.String()] = sum

	return cnab.BundleReference{Reference: ref}, nil
}

func (t *TestRegistry) PushBundle(ctx context.Context, ref cnab.BundleReference, opts RegistryOptions) (cnab.BundleReference, error) {
	if t.MockPushBundle != nil {
		return t.MockPushBundle(ctx, ref, opts)
	}

	return ref, nil
}

func (t *TestRegistry) PushImage(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (digest.Digest, error) {
	if t.MockPushImage != nil {
		return t.MockPushImage(ctx, ref, opts)
	}
	return "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041", nil
}

func (t *TestRegistry) GetCachedImage(ctx context.Context, ref cnab.OCIReference) (ImageMetadata, error) {
	if t.MockGetCachedImage != nil {
		return t.MockGetCachedImage(ctx, ref)
	}

	img := ref.String()
	sum, ok := t.cache[img]
	if !ok {
		return ImageMetadata{}, fmt.Errorf("failed to find image in docker cache: %w", ErrNotFound{Reference: ref})
	}
	return sum, nil
}

func (t TestRegistry) GetImageMetadata(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (ImageMetadata, error) {
	meta, err := t.GetCachedImage(ctx, ref)
	if err != nil {
		if t.MockGetImageMetadata != nil {
			return t.MockGetImageMetadata(ctx, ref, opts)
		} else {
			mockImg := ImageMetadata{
				Reference:   ref,
				RepoDigests: []string{fmt.Sprintf("%s@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041", ref.Repository())},
			}
			return mockImg, nil
		}
	}
	return meta, nil
}

func (t *TestRegistry) ListTags(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) ([]string, error) {
	if t.MockListTags != nil {
		return t.MockListTags(ctx, ref, opts)
	}

	return nil, nil
}

func (t *TestRegistry) PullImage(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) error {
	if t.MockPullImage != nil {
		return t.MockPullImage(ctx, ref, opts)
	}

	image := ref.String()
	sum, err := NewImageSummaryFromInspect(ref, types.ImageInspect{
		ID:          cnab.NewULID(),
		RepoDigests: []string{fmt.Sprintf("%s@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041", image)},
	})
	if err != nil {
		return err
	}
	t.cache[image] = sum
	return nil
}

func (t TestRegistry) GetBundleMetadata(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (BundleMetadata, error) {
	if t.MockGetBundleMetadata != nil {
		return t.MockGetBundleMetadata(ctx, ref, opts)
	}

	return BundleMetadata{}, ErrNotFound{Reference: ref}
}
