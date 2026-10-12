package client

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/test"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
)

func TestRunner_Validate(t *testing.T) {
	r := NewTestRunner(t, "lucky-charms", "cereals", true)

	err := r.Validate()
	require.NoError(t, err)
}

func TestRunner_Validate_MissingName(t *testing.T) {
	// Setup failure: empty package name
	r := NewTestRunner(t, "", "candy", true)

	err := r.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "package name to execute not specified")
}

func TestRunner_Validate_MissingExecutable(t *testing.T) {
	r := NewTestRunner(t, "mypackage", "packages", true)

	// Setup failure: Don't copy the package binary into the test context
	err := r.FileSystem.Remove(r.getExecutablePath())
	require.NoError(t, err)

	err = r.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "package not found")
}

func TestRunner_Run_CensorsError(t *testing.T) {
	r := NewTestRunner(t, "mypackage", "mixins", true)
	r.TestContext.Setenv(test.ExpectedCommandExitCodeEnv, "1")
	r.TestContext.Setenv(test.ExpectedCommandErrorEnv, "couldn't run command fail open_door topsecret")
	r.SetSensitiveValues([]string{"topsecret"})

	err := r.Run(context.Background(), pkgmgmt.CommandOptions{Command: "install", Runtime: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "couldn't run command fail open_door *******")
	assert.NotContains(t, err.Error(), "topsecret", "expected the sensitive value to be masked in the error")
	assert.NotContains(t, r.TestContext.GetError(), "topsecret", "expected the sensitive value to be masked in the output of the command")
}

func TestRunner_Run_ErrorIncludesRunnableCommand(t *testing.T) {
	r := NewTestRunner(t, "mypackage", "mixins", true)
	r.TestContext.Setenv(test.ExpectedCommandExitCodeEnv, "1")

	err := r.Run(context.Background(), pkgmgmt.CommandOptions{Command: "install", Runtime: true})
	require.Error(t, err)
	// The test command is executed with the test binary
	testBin := portercontext.FormatCommand(&exec.Cmd{Args: []string{os.Args[0]}})
	assert.Contains(t, err.Error(), "package command failed "+testBin+" ", "expected the error to include a runnable command")
	assert.NotContains(t, err.Error(), r.Getwd()+os.Args[0], "expected the command to not be prefixed with the working directory")
}

func TestRunner_Run_ErrorQuotesCommandArguments(t *testing.T) {
	r := NewTestRunner(t, "mypackage", "mixins", true)
	r.TestContext.Setenv(test.ExpectedCommandExitCodeEnv, "1")

	err := r.Run(context.Background(), pkgmgmt.CommandOptions{Command: "install", File: "my file.yaml", Runtime: true})
	require.Error(t, err)
	wantArgs := "install -f 'my file.yaml'"
	if runtime.GOOS == "windows" {
		wantArgs = `install -f "my file.yaml"`
	}
	assert.Contains(t, err.Error(), wantArgs, "expected arguments with spaces to be quoted so that the command is runnable")
}

func TestRunner_Run_TracesRunnableCommand(t *testing.T) {
	// Record the spans created by the runner
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	tracer := tracing.NewTracer(provider.Tracer(t.Name()), nil)
	ctx, rootSpan := tracer.Start(context.Background(), t.Name())
	ctx, log := tracing.NewRootLogger(ctx, rootSpan, zap.NewNop(), tracer)

	r := NewTestRunner(t, "mypackage", "mixins", true)
	r.TestContext.Setenv(test.ExpectedCommandExitCodeEnv, "1")

	err := r.Run(ctx, pkgmgmt.CommandOptions{Command: "install", File: "my file.yaml", Runtime: true})
	require.Error(t, err)
	log.EndSpan()

	attrs := map[attribute.Key]string{}
	for _, span := range recorder.Ended() {
		for _, attr := range span.Attributes() {
			attrs[attr.Key] = attr.Value.AsString()
		}
	}
	// The test command is executed with the test binary
	wantCmd := portercontext.FormatCommand(&exec.Cmd{
		Args: []string{os.Args[0], r.getExecutablePath(), "install", "-f", "my file.yaml"},
	})
	assert.Equal(t, wantCmd, attrs["command"], "expected the command attribute to be a runnable command")
	assert.Equal(t, r.Getwd(), attrs["dir"], "expected the working directory to be recorded separately from the command")
}
