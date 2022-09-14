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
	cxt.Setenv("PORTER_DEBUG", "true")

	mixins := mixin.NewTestMixinProvider()
	cfg := NewConfigFor(cxt.Context)
	pr := NewPorterRuntime(cfg, mixins)

	return &TestPorterRuntime{
		TestContext:   cxt,
		PorterRuntime: pr,
	}
}

type TestRuntimeConfig struct {
	RuntimeConfig
	TestContext *portercontext.TestContext
}

func NewTestRuntimeConfig(t *testing.T) TestRuntimeConfig {
	porterCtx := portercontext.NewTestContext(t)
	porterCtx.Setenv("PORTER_DEBUG", "true")
	runtimeCfg := NewConfigFor(porterCtx.Context)
	return TestRuntimeConfig{
		RuntimeConfig: runtimeCfg,
		TestContext:   porterCtx,
	}
}
