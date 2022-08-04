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
	MockPullBundle          func(ref cnab.OCIReference, insecureRegistry bool) (cnab.BundleReference, error)
	MockPushBundle          func(bundleRef cnab.BundleReference, insecureRegistry bool) (bundleReference cnab.BundleReference, err error)
	MockPushInvocationImage func(ctx context.Context, invocationImage string) (imageDigest digest.Digest, err error)
	MockGetCachedImage      func(ctx context.Context, invocationImage string) (ImageSummary, error)
	MockListTags            func(ctx context.Context, repository string) ([]string, error)
	MockPullImage           func(ctx context.Context, image string) error
	cache                   map[string]ImageSummary
}

func NewTestRegistry() *TestRegistry {
	return &TestRegistry{
		cache: make(map[string]ImageSummary),
	}
}

func (t TestRegistry) PullBundle(ctx context.Context, ref cnab.OCIReference, insecureRegistry bool) (cnab.BundleReference, error) {
	if t.MockPullBundle != nil {
		return t.MockPullBundle(ref, insecureRegistry)
	}
	sum, _ := NewImageSummary(ref.String(), types.ImageInspect{ID: cnab.NewULID()})
	t.cache[ref.String()] = sum

	return cnab.BundleReference{Reference: ref}, nil
}

func (t *TestRegistry) PushBundle(ctx context.Context, bundleRef cnab.BundleReference, insecureRegistry bool) (cnab.BundleReference, error) {
	if t.MockPushBundle != nil {
		return t.MockPushBundle(bundleRef, insecureRegistry)
	}

	return bundleRef, nil
}

func (t *TestRegistry) PushInvocationImage(ctx context.Context, invocationImage string) (digest.Digest, error) {
	if t.MockPushInvocationImage != nil {
		return t.MockPushInvocationImage(ctx, invocationImage)
	}
	return "", nil
}

func (t *TestRegistry) GetCachedImage(ctx context.Context, img string) (ImageSummary, error) {
	if t.MockGetCachedImage != nil {
		return t.MockGetCachedImage(ctx, img)
	}

	sum, ok := t.cache[img]
	if !ok {
		return ImageSummary{}, fmt.Errorf("image %s not found in cache", img)
	}
	return sum, nil
}

func (t *TestRegistry) ListTags(ctx context.Context, repository string) ([]string, error) {
	if t.MockListTags != nil {
		return t.MockListTags(ctx, repository)
	}

	return nil, nil
}

func (t *TestRegistry) PullImage(ctx context.Context, image string) error {
	if t.MockPullImage != nil {
		return t.MockPullImage(ctx, image)
	}
	sum, err := NewImageSummary(image, types.ImageInspect{
		ID:          cnab.NewULID(),
		RepoDigests: []string{fmt.Sprintf("%s@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041", image)},
	})
	if err != nil {
		return err
	}
	t.cache[image] = sum
	return nil
}
