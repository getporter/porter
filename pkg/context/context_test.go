package context

import (
	"context"
	"errors"
	"testing"

	"get.porter.sh/porter/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestContext_EnvironMap(t *testing.T) {
	c := NewTestContext(t)
	c.Clearenv()

	c.Setenv("a", "1")
	c.Setenv("b", "2")

	got := c.EnvironMap()

	want := map[string]string{
		"a": "1",
		"b": "2",
	}
	assert.Equal(t, want, got)

	// Make sure we have a copy
	got["c"] = "3"
	assert.Empty(t, c.Getenv("c"), "Expected to get a copy of the context's environment variables")
}

func TestContext_LogToFile(t *testing.T) {
	c := NewTestContext(t)
	c.ConfigureLogging(LogConfiguration{LogLevel: zapcore.DebugLevel, LogToFile: true, LogDirectory: "/.porter/logs"})
	c.timestampLogs = false // turn off timestamps so we can compare more easily
	logfile := c.logFile.Name()
	_, log := c.StartRootSpan(context.Background(), t.Name())
	log.Info("a thing happened")
	log.Warn("a weird thing happened")
	log.Error(errors.New("a bad thing happened"))
	log.EndSpan()
	c.Close()

	// Check that the logs are in json
	logContents, err := c.FileSystem.ReadFile(logfile)
	require.NoError(t, err)
	c.CompareGoldenFile("testdata/expected-logs.txt", string(logContents))

	// Compare the human readable logs sent to stderr
	c.CompareGoldenFile("testdata/expected-output.txt", c.GetError())
}

func TestContext_UserAgent(t *testing.T) {
	t.Run("append versions when available", func(t *testing.T) {
		pkg.Version = "v1.0.0"
		pkg.Commit = "abc123"
		c := NewTestContext(t)

		require.Contains(t, c.UserAgent(), "porter/"+pkg.Version)
	})

	t.Run("append commit hash when version is not available", func(t *testing.T) {
		pkg.Version = ""
		pkg.Commit = "abc123"
		c := NewTestContext(t)

		require.Contains(t, c.UserAgent(), "porter/"+pkg.Commit)
	})

	t.Run("omit slash when neither version nor commit hash is available", func(t *testing.T) {
		pkg.Version = ""
		pkg.Commit = ""
		c := NewTestContext(t)

		require.Contains(t, c.UserAgent(), "porter")
	})
}
