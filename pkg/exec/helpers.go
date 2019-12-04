package exec

import (
	"testing"

	"get.porter.sh/porter/pkg/context"
)

type TestMixin struct {
	*Mixin
	TestContext *context.TestContext
}

// NewTestMixin initializes a helm mixin, with the output buffered, and an in-memory file system.
func NewTestMixin(t *testing.T) *TestMixin {
	tc := context.NewTestContext(t)
	m := New()
	m.Context = tc.Context
	return &TestMixin{
		Mixin:       m,
		TestContext: tc,
	}
}
