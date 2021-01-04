package cnabtooci

import (
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/pkg/errors"
)

var _ RegistryProvider = &TestRegistry{}

type TestRegistry struct {
	MockPullBundle          func(tag string, insecureRegistry bool) (bun bundle.Bundle, reloMap *relocation.ImageRelocationMap, err error)
	MockPushBundle          func(bun bundle.Bundle, tag string, insecureRegistry bool) (reloMap *relocation.ImageRelocationMap, err error)
	MockPushInvocationImage func(invocationImage string) (imageDigest string, err error)
}

func NewTestRegistry() *TestRegistry {
	return &TestRegistry{}
}

func (t TestRegistry) PullBundle(tag string, insecureRegistry bool) (bundle.Bundle, *relocation.ImageRelocationMap, error) {
	if t.MockPullBundle != nil {
		return t.MockPullBundle(tag, insecureRegistry)
	}

	return bundle.Bundle{}, nil, errors.Errorf("tried to pull %s but MockPullBundle was not set", tag)
}

func (t TestRegistry) PushBundle(bun bundle.Bundle, tag string, insecureRegistry bool) (*relocation.ImageRelocationMap, error) {
	if t.MockPushBundle != nil {
		return t.MockPushBundle(bun, tag, insecureRegistry)
	}

	return nil, nil
}

func (t TestRegistry) PushInvocationImage(invocationImage string) (string, error) {
	if t.MockPushInvocationImage != nil {
		return t.MockPushInvocationImage(invocationImage)
	}
	return "", nil
}
