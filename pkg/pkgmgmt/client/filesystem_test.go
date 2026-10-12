package client

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestFileSystem_List(t *testing.T) {
	c := config.NewTestConfig(t)

	p := NewFileSystem(c.Config, "mixins")
	mixins, err := p.List()

	require.Nil(t, err)
	require.Len(t, mixins, 2)
	assert.Equal(t, mixins[0], "exec")
	assert.Equal(t, mixins[1], "testmixin")
}

func TestFileSystem_GetMetadata_CommandFails(t *testing.T) {
	const cmdErr = "unknown command version"

	testcases := []struct {
		name       string
		verbosity  zapcore.Level
		wantStderr bool
	}{
		{name: "debug", verbosity: zapcore.DebugLevel, wantStderr: true},
		{name: "info", verbosity: zapcore.InfoLevel, wantStderr: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := config.NewTestConfig(t)
			c.TestContext.Setenv(test.ExpectedCommandExitCodeEnv, "1")
			c.TestContext.Setenv(test.ExpectedCommandErrorEnv, cmdErr)
			c.ConfigureLogging(context.Background(), portercontext.LogConfiguration{Verbosity: tc.verbosity})
			ctx, log := c.StartRootSpan(context.Background(), t.Name())
			defer log.Close()

			p := NewFileSystem(c.Config, "mixins")
			_, err := p.GetMetadata(ctx, "exec")

			require.Error(t, err)
			assert.Contains(t, err.Error(), cmdErr, "expected the error to include the output of the failed command")
			if tc.wantStderr {
				assert.Contains(t, c.TestContext.GetError(), cmdErr, "expected the output of the failed command to be printed when verbosity is debug")
			} else {
				assert.NotContains(t, c.TestContext.GetError(), cmdErr, "expected the output of the failed command to not be printed when verbosity is not debug")
			}
		})
	}
}
