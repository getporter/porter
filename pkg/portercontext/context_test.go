package portercontext

import (
	"context"
	"errors"
	"testing"

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
	c.ConfigureLogging(context.Background(), LogConfiguration{
		Verbosity:    zapcore.DebugLevel,
		LogLevel:     zapcore.DebugLevel,
		LogToFile:    true,
		LogDirectory: "/.porter/logs",
	})
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
