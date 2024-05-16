package runtime

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/portercontext"
)

type TestPorterRuntime struct {
	*PorterRuntime
	TestContext *portercontext.TestContext
}

func NewTestPorterRuntime(t *testing.T) *TestPorterRuntime {
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("PORTER_DEBUG", "true")

	mixins := mixin.NewTestMixinProvider()
	cfg := NewConfigFor(testConfig.Config)
	pr := NewPorterRuntime(cfg, mixins)

	return &TestPorterRuntime{
		TestContext:   testConfig.TestContext,
		PorterRuntime: pr,
	}
}

type TestRuntimeConfig struct {
	RuntimeConfig
	TestContext *portercontext.TestContext
}

func NewTestRuntimeConfig(t *testing.T) TestRuntimeConfig {
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("PORTER_DEBUG", "true")
	runtimeCfg := NewConfigFor(testConfig.Config)
	return TestRuntimeConfig{
		RuntimeConfig: runtimeCfg,
		TestContext:   testConfig.TestContext,
	}
}
