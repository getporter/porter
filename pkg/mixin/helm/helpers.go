package helm

import (
	"testing"

	"github.com/deislabs/porter/pkg/context"

	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
)

type TestMixin struct {
	*Mixin
	TestContext *context.TestContext
}

type testKubernetesFactory struct {
}

func (t *testKubernetesFactory) GetClient(configPath string) (kubernetes.Interface, error) {
	return testclient.NewSimpleClientset(), nil
}

// NewTestMixin initializes a helm mixin, with the output buffered, and an in-memory file system.
func NewTestMixin(t *testing.T) *TestMixin {
	c := context.NewTestContext(t)
	m := &TestMixin{
		Mixin: &Mixin{
			Context:       c.Context,
			ClientFactory: &testKubernetesFactory{},
		},
		TestContext: c,
	}

	return m
}
