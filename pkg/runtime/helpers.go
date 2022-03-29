package runtime

import (
	"testing"

	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/portercontext"
)

type TestPorterRuntime struct {
	*PorterRuntime
	TestContext *portercontext.TestContext
}

func NewTestPorterRuntime(t *testing.T) *TestPorterRuntime {
	cxt := portercontext.NewTestContext(t)
	mixins := mixin.NewTestMixinProvider()
	pr := NewPorterRuntime(cxt.Context, mixins)

	return &TestPorterRuntime{
		TestContext:   cxt,
		PorterRuntime: pr,
	}
}
