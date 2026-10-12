package plugins

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"get.porter.sh/porter/pkg/config"
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

func TestPluginRunner_Run_TracesRunnableCommand(t *testing.T) {
	// Record the spans created by the runner
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	tracer := tracing.NewTracer(provider.Tracer(t.Name()), nil)
	ctx, rootSpan := tracer.Start(context.Background(), t.Name())
	ctx, log := tracing.NewRootLogger(ctx, rootSpan, zap.NewNop(), tracer)

	t.Setenv(config.EnvHOME, "/home/myuser/.porter")
	c := portercontext.NewTestContext(t)
	c.Setenv(test.ExpectedCommandExitCodeEnv, "1")
	r := NewRunner("myplugin")
	r.Context = c.Context

	err := r.Run(ctx, CommandOptions{Command: "version --output json"})
	require.Error(t, err)
	log.EndSpan()

	attrs := map[attribute.Key]string{}
	for _, span := range recorder.Ended() {
		for _, attr := range span.Attributes() {
			attrs[attr.Key] = attr.Value.AsString()
		}
	}
	pluginPath, err := config.New().GetPluginPath("myplugin")
	require.NoError(t, err)
	// The test command is executed with the test binary
	wantCmd := portercontext.FormatCommand(&exec.Cmd{
		Args: []string{os.Args[0], pluginPath, "version", "--output", "json"},
	})
	assert.Equal(t, wantCmd, attrs["full-command"], "expected the full-command attribute to be a runnable command")
	assert.Equal(t, r.Getwd(), attrs["dir"], "expected the working directory to be recorded separately from the command")
}
