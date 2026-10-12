package tracing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

func newTestTraceLogger() traceLogger {
	tracer := noop.NewTracerProvider().Tracer("noop")
	ctx, span := tracer.Start(context.Background(), "test")
	return newTraceLogger(ctx, span, zap.NewNop(), NewTracer(tracer, nil))
}

func innerFailure(log TraceLogger) error {
	return log.Error(io.ErrUnexpectedEOF)
}

func outerFailure(log TraceLogger) error {
	err := innerFailure(log)
	return log.Error(fmt.Errorf("outer failed: %w", err))
}

func TestStackTrace(t *testing.T) {
	t.Run("error without a stack trace", func(t *testing.T) {
		assert.Empty(t, StackTrace(nil))
		assert.Empty(t, StackTrace(errors.New("plain")))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.NoError(t, newTestTraceLogger().Error(nil))
	})

	t.Run("error is unchanged", func(t *testing.T) {
		err := innerFailure(newTestTraceLogger())
		assert.EqualError(t, err, io.ErrUnexpectedEOF.Error())
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})

	t.Run("readable stack trace", func(t *testing.T) {
		err := innerFailure(newTestTraceLogger())

		lines := strings.Split(StackTrace(err), "\n")
		require.Greater(t, len(lines), 2)
		assert.Equal(t, "Stack trace:", lines[0])
		assert.Regexp(t, `^  at pkg/tracing\.innerFailure \(pkg/tracing/stack_test\.go:\d+\)$`, lines[1],
			"the stack trace should start where the error was recorded")
		assert.NotContains(t, StackTrace(err), "traceLogger", "the logger should not be in the stack trace")
		assert.NotContains(t, StackTrace(err), "runtime.", "the go runtime should not be in the stack trace")
	})

	t.Run("Errorf", func(t *testing.T) {
		err := newTestTraceLogger().Errorf("failed: %w", io.ErrUnexpectedEOF)
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
		assert.Contains(t, StackTrace(err), "\n  at pkg/tracing.TestStackTrace.func")
		assert.NotContains(t, StackTrace(err), "traceLogger")
	})

	t.Run("innermost stack trace is kept", func(t *testing.T) {
		err := outerFailure(newTestTraceLogger())
		assert.EqualError(t, err, "outer failed: unexpected EOF")
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)

		stack := StackTrace(err)
		assert.Contains(t, stack, "\n  at pkg/tracing.innerFailure (")
		assert.Contains(t, stack, "\n  at pkg/tracing.outerFailure (")
		assert.Equal(t, 1, strings.Count(stack, "Stack trace:"))
	})
}

func TestTraceLogger_Error_RecordsStackTrace(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	tracer := provider.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test")
	log := newTraceLogger(ctx, span, zap.NewNop(), NewTracer(tracer, nil))

	// Record the same error twice, like when it is returned through multiple spans
	err := innerFailure(log)
	_ = log.Error(err)
	span.End()

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	events := spans[0].Events()
	require.Len(t, events, 2)
	for _, event := range events {
		attrs := make(map[string]string, len(event.Attributes))
		for _, attr := range event.Attributes {
			attrs[string(attr.Key)] = attr.Value.String()
		}
		assert.Equal(t, "*errors.errorString", attrs["exception.type"], "the original error type should be recorded")
		stack := attrs["exception.stacktrace"]
		assert.Regexp(t, `^get\.porter\.sh/porter/pkg/tracing\.innerFailure\n\t.*stack_test\.go:\d+\n`, stack,
			"the stack trace should use the format of the go runtime and start where the error was recorded")
		assert.NotContains(t, stack, "Stack trace:", "the console format should not be sent to telemetry")
		assert.NotContains(t, stack, "traceLogger", "the logger should not be in the stack trace")
	}
}

func TestTelemetryStackTrace_NoStack(t *testing.T) {
	assert.Empty(t, telemetryStackTrace(nil))
	assert.Empty(t, telemetryStackTrace(errors.New("plain")))
	assert.Empty(t, telemetryStackTrace(&stackError{err: errors.New("no frames")}))
	assert.Empty(t, StackTrace(&stackError{err: errors.New("no frames")}))
}

func TestFormatFile(t *testing.T) {
	assert.Equal(t, "pkg/porter/show.go", formatFile(moduleRoot+"pkg/porter/show.go"))
	assert.Equal(t, "github.com/spf13/cobra@v1.9.1/command.go", formatFile("/home/me/go/pkg/mod/github.com/spf13/cobra@v1.9.1/command.go"))
	assert.Equal(t, "/usr/local/go/src/testing/testing.go", formatFile("/usr/local/go/src/testing/testing.go"))
}
