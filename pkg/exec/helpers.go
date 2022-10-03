package exec

import (
	"testing"

	"get.porter.sh/porter/pkg/runtime"
)

type TestMixin struct {
	*Mixin
	TestConfig runtime.TestRuntimeConfig
}

// NewTestMixin initializes an exec mixin, with the output buffered, and an in-memory file system.
func NewTestMixin(t *testing.T) *TestMixin {
	cfg := runtime.NewTestRuntimeConfig(t)
	m := New()
	m.Config = cfg.RuntimeConfig
	return &TestMixin{
		Mixin:      m,
		TestConfig: cfg,
	}
}
