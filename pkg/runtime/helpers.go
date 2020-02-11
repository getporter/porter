package runtime

import (
	"testing"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/mixin"
)

type TestPorterRuntime struct {
	*PorterRuntime
	TestContext *context.TestContext
}

func NewTestPorterRuntime(t *testing.T) *TestPorterRuntime {
	cxt := context.NewTestContext(t)
	mixins := mixin.NewTestMixinProvider()
	pr := NewPorterRuntime(cxt.Context, mixins)

	return &TestPorterRuntime{
		TestContext:   cxt,
		PorterRuntime: pr,
	}
}
