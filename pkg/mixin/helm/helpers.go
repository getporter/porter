package helm

import (
	"testing"

	"github.com/deislabs/porter/pkg/context"
)

type TestMixin struct {
	*Mixin
	TestContext *context.TestContext
}

// NewTestMixin initializes a helm mixin, with the output buffered, and an in-memory file system.
func NewTestMixin(t *testing.T) *TestMixin {
	c := context.NewTestContext(t)
	m := &TestMixin{
		Mixin: &Mixin{
			Context: c.Context,
		},
		TestContext: c,
	}

	return m
}
