package kubernetes

import (
	"testing"

	"github.com/deislabs/porter/pkg/context"
)

type TestMixin struct {
	*Mixin
	TestContext *context.TestContext
}

func NewTestMixin(t *testing.T) *TestMixin {
	c := context.NewTestContext(t)
	m := New()
	m.Context = c.Context
	return &TestMixin{
		Mixin:       m,
		TestContext: c,
	}
}
