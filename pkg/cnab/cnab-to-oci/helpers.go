package cnabtooci

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/opencontainers/go-digest"
)

var _ RegistryProvider = &TestRegistry{}

type TestRegistry struct {
	MockPullBundle          func(ref cnab.OCIReference, insecureRegistry bool) (cnab.BundleReference, error)
	MockPushBundle          func(bundleRef cnab.BundleReference, insecureRegistry bool) (bundleReference cnab.BundleReference, err error)
	MockPushInvocationImage func(ctx context.Context, invocationImage string) (imageDigest digest.Digest, err error)
	MockIsImageCached       func(ctx context.Context, invocationImage string) (bool, error)
	MockListTags            func(ctx context.Context, repository string) ([]string, error)
	MockPullImage           func(ctx context.Context, image string) (string, error)
}

func NewTestRegistry() *TestRegistry {
	return &TestRegistry{}
}

func (t TestRegistry) PullBundle(ctx context.Context, ref cnab.OCIReference, insecureRegistry bool) (cnab.BundleReference, error) {
	if t.MockPullBundle != nil {
		return t.MockPullBundle(ref, insecureRegistry)
	}

	return cnab.BundleReference{Reference: ref}, nil
}

func (t TestRegistry) PushBundle(ctx context.Context, bundleRef cnab.BundleReference, insecureRegistry bool) (cnab.BundleReference, error) {
	if t.MockPushBundle != nil {
		return t.MockPushBundle(bundleRef, insecureRegistry)
	}

	return bundleRef, nil
}

func (t TestRegistry) PushInvocationImage(ctx context.Context, invocationImage string) (digest.Digest, error) {
	if t.MockPushInvocationImage != nil {
		return t.MockPushInvocationImage(ctx, invocationImage)
	}
	return "", nil
}

func (t TestRegistry) IsImageCached(ctx context.Context, invocationImage string) (bool, error) {
	if t.MockIsImageCached != nil {
		return t.MockIsImageCached(ctx, invocationImage)
	}

	return true, nil
}

func (t TestRegistry) ListTags(ctx context.Context, repository string) ([]string, error) {
	if t.MockListTags != nil {
		return t.MockListTags(ctx, repository)
	}

	return nil, nil
}

func (t TestRegistry) PullImage(ctx context.Context, image string) (string, error) {
	if t.MockPullImage != nil {
		return t.MockPullImage(ctx, image)
	}
	return "", nil
}
