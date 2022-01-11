package cnabtooci

import (
	"get.porter.sh/porter/pkg/cnab"
	"github.com/opencontainers/go-digest"
)

var _ RegistryProvider = &TestRegistry{}

type TestRegistry struct {
	MockPullBundle          func(ref cnab.OCIReference, insecureRegistry bool) (cnab.BundleReference, error)
	MockPushBundle          func(bundleRef cnab.BundleReference, insecureRegistry bool) (bundleReference cnab.BundleReference, err error)
	MockPushInvocationImage func(invocationImage string) (imageDigest digest.Digest, err error)
	MockIsImageCached       func(invocationImage string) (bool, error)
}

func NewTestRegistry() *TestRegistry {
	return &TestRegistry{}
}

func (t TestRegistry) PullBundle(ref cnab.OCIReference, insecureRegistry bool) (cnab.BundleReference, error) {
	if t.MockPullBundle != nil {
		return t.MockPullBundle(ref, insecureRegistry)
	}

	return cnab.BundleReference{Reference: ref}, nil
}

func (t TestRegistry) PushBundle(bundleRef cnab.BundleReference, insecureRegistry bool) (cnab.BundleReference, error) {
	if t.MockPushBundle != nil {
		return t.MockPushBundle(bundleRef, insecureRegistry)
	}

	return bundleRef, nil
}

func (t TestRegistry) PushInvocationImage(invocationImage string) (digest.Digest, error) {
	if t.MockPushInvocationImage != nil {
		return t.MockPushInvocationImage(invocationImage)
	}
	return "", nil
}

func (t TestRegistry) IsImageCached(invocationImage string) (bool, error) {
	if t.MockIsImageCached != nil {
		return t.MockIsImageCached(invocationImage)
	}

	return true, nil
}
