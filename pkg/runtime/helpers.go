package runtime

import (
	"testing"

	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/mixin"
)

type TestPorterRuntime struct {
	*PorterRuntime
	TestContext *context.TestContext
}

func NewTestPorterRuntime(t *testing.T) *TestPorterRuntime {
	cxt := context.NewTestContext(t)
	mixins := &mixin.TestMixinProvider{}
	pr := NewPorterRuntime(cxt.Context, mixins)

	return &TestPorterRuntime{
		TestContext:   cxt,
		PorterRuntime: pr,
	}
}
