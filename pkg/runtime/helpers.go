package runtime

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/mixin"
)

type TestPorterRuntime struct {
	*PorterRuntime
	TestContext *context.TestContext
}

func NewTestPorterRuntime(t *testing.T) *TestPorterRuntime {
	c := config.NewTestConfig(t)
	mixins := mixin.NewTestMixinProvider(c)
	pr := NewPorterRuntime(c.Context, mixins)

	return &TestPorterRuntime{
		TestContext:   c.TestContext,
		PorterRuntime: pr,
	}
}
