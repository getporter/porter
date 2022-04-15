package exec

import (
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
)

type TestMixin struct {
	*Mixin
	TestContext *portercontext.TestContext
}

// NewTestMixin initializes an exec mixin, with the output buffered, and an in-memory file system.
func NewTestMixin(t *testing.T) *TestMixin {
	tc := portercontext.NewTestContext(t)
	m := New()
	m.Context = tc.Context
	return &TestMixin{
		Mixin:       m,
		TestContext: tc,
	}
}
