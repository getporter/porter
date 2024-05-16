package runtime

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeConfig_DebugMode(t *testing.T) {
	testcases := []struct {
		debugEnv  string
		wantDebug bool
	}{
		{debugEnv: "true", wantDebug: true},
		{debugEnv: "1", wantDebug: true},
		{debugEnv: "abc", wantDebug: false},
		{debugEnv: "", wantDebug: false},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.debugEnv, func(t *testing.T) {
			t.Parallel()

			config := config.NewTestConfig(t)
			config.Setenv("PORTER_DEBUG", tc.debugEnv)
			c := NewConfigFor(config.Config)
			assert.Equal(t, tc.wantDebug, c.DebugMode)
		})
	}
}

func TestNewTestPorterRuntime(t *testing.T) {
	r := NewTestPorterRuntime(t)
	assert.True(t, r.config.DebugMode)
}

func TestNewTestRuntimeConfig(t *testing.T) {
	cfg := NewTestRuntimeConfig(t)
	assert.True(t, cfg.DebugMode)
}
