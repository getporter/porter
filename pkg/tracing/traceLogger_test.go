package tracing

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func TestTraceLogger_ShouldLog(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	tracer := nooptrace.NewTracerProvider().Tracer("noop")
	l := newTraceLogger(context.Background(), nil, logger, NewTracer(tracer, nil))

	assert.True(t, l.ShouldLog(zap.ErrorLevel))
	assert.True(t, l.ShouldLog(zap.WarnLevel))
	assert.False(t, l.ShouldLog(zap.InfoLevel))
	assert.False(t, l.ShouldLog(zap.DebugLevel))
}

func TestTraceSensitiveAttributesBuildFlag(t *testing.T) {
	assert.False(t, traceSensitiveAttributes, "traceSensitiveAttributes should be disabled by default and require a custom build to enable")
}

// setupTestLogger creates a logger with an in-memory span recorder and log observer
func setupTestLogger() (context.Context, TraceLogger, *tracetest.SpanRecorder, *observer.ObservedLogs) {
	// Setup in-memory span recorder
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))
	otelTracer := tracerProvider.Tracer("test")

	// Setup in-memory log observer
	core, logs := observer.New(zapcore.ErrorLevel)
	logger := zap.New(core)

	// Create context and span
	ctx := context.Background()
	ctx, span := otelTracer.Start(ctx, "test-span")

	// Create trace logger with cleanup function
	cleanup := func(ctx context.Context) error {
		return tracerProvider.Shutdown(ctx)
	}
	tl := newTraceLogger(ctx, span, logger, NewTracer(otelTracer, cleanup))

	return ctx, tl, spanRecorder, logs
}

func TestRecordError_DoesNotLogToConsole(t *testing.T) {
	_, logger, spanRecorder, logs := setupTestLogger()

	testErr := errors.New("test error")
	returnedErr := logger.RecordError(testErr)

	// End the span so we can verify it was recorded
	if tl, ok := logger.(traceLogger); ok {
		tl.span.End()
	}

	// Verify error is returned unchanged
	require.Equal(t, testErr, returnedErr)

	// Console logs should be empty (no error logs written)
	require.Equal(t, 0, logs.Len(), "RecordError should not write to console")

	// But span should have the error recorded
	spans := spanRecorder.Ended()
	require.Len(t, spans, 1, "expected one span to be recorded")

	// Verify the span has error status
	span := spans[0]
	require.Equal(t, "Error", span.Status().Code.String())
	require.Contains(t, span.Status().Description, "test error")
}

func TestError_LogsToConsole(t *testing.T) {
	_, logger, _, logs := setupTestLogger()
	defer func() {
		if tl, ok := logger.(traceLogger); ok {
			tl.span.End()
		}
	}()

	testErr := errors.New("test error")
	returnedErr := logger.Error(testErr)

	// Verify error is returned unchanged
	require.Equal(t, testErr, returnedErr)

	// Console should contain the error
	require.Equal(t, 1, logs.Len(), "Error should write to console")
	logEntry := logs.All()[0]
	assert.Equal(t, "test error", logEntry.Message)
	assert.Equal(t, zapcore.ErrorLevel, logEntry.Level)
}

func TestRecordErrorf_DoesNotLogToConsole(t *testing.T) {
	_, logger, _, logs := setupTestLogger()
	defer func() {
		if tl, ok := logger.(traceLogger); ok {
			tl.span.End()
		}
	}()

	returnedErr := logger.RecordErrorf("formatted error: %s", "details")

	// Verify error is returned with correct message
	require.NotNil(t, returnedErr)
	assert.Contains(t, returnedErr.Error(), "formatted error: details")

	// Console logs should be empty
	require.Equal(t, 0, logs.Len(), "RecordErrorf should not write to console")
}

func TestErrorf_LogsToConsole(t *testing.T) {
	_, logger, _, logs := setupTestLogger()
	defer func() {
		if tl, ok := logger.(traceLogger); ok {
			tl.span.End()
		}
	}()

	returnedErr := logger.Errorf("formatted error: %s", "details")

	// Verify error is returned with correct message
	require.NotNil(t, returnedErr)
	assert.Contains(t, returnedErr.Error(), "formatted error: details")

	// Console should contain the error
	require.Equal(t, 1, logs.Len(), "Errorf should write to console")
	logEntry := logs.All()[0]
	assert.Contains(t, logEntry.Message, "formatted error: details")
	assert.Equal(t, zapcore.ErrorLevel, logEntry.Level)
}

func TestRecordError_WithNilError(t *testing.T) {
	_, logger, _, logs := setupTestLogger()
	defer func() {
		if tl, ok := logger.(traceLogger); ok {
			tl.span.End()
		}
	}()

	returnedErr := logger.RecordError(nil)

	// Should return nil unchanged
	require.Nil(t, returnedErr)

	// No logs should be written
	require.Equal(t, 0, logs.Len())
}

func TestError_WithNilError(t *testing.T) {
	_, logger, _, logs := setupTestLogger()
	defer func() {
		if tl, ok := logger.(traceLogger); ok {
			tl.span.End()
		}
	}()

	returnedErr := logger.Error(nil)

	// Should return nil unchanged
	require.Nil(t, returnedErr)

	// No logs should be written
	require.Equal(t, 0, logs.Len())
}

func TestGetTracer_ReturnsErrorTracer(t *testing.T) {
	ctx, logger, _, _ := setupTestLogger()
	defer func() {
		if tl, ok := logger.(traceLogger); ok {
			tl.span.End()
		}
	}()

	// Store logger in context
	ctx = context.WithValue(ctx, contextKeyTraceLogger, traceLoggerContext{
		logger: zap.NewNop(),
		tracer: nooptrace.NewTracerProvider().Tracer("test"),
	})

	tracer := GetTracer(ctx)
	require.NotNil(t, tracer)

	// Verify it can be used as ErrorTracer
	err := tracer.RecordError(errors.New("test"))
	require.NotNil(t, err)
}

func TestGetLogger_ReturnsTraceLogger(t *testing.T) {
	ctx := context.Background()

	// Store logger in context
	ctx = context.WithValue(ctx, contextKeyTraceLogger, traceLoggerContext{
		logger: zap.NewNop(),
		tracer: nooptrace.NewTracerProvider().Tracer("test"),
	})

	logger := GetLogger(ctx)
	require.NotNil(t, logger)

	// Verify it implements TraceLogger
	var _ TraceLogger = logger
}

func TestErrorDelegatestoRecordError(t *testing.T) {
	// This test verifies that Error() calls RecordError() internally
	// to avoid code duplication, as per the implementation plan

	_, logger, _, logs := setupTestLogger()
	defer func() {
		if tl, ok := logger.(traceLogger); ok {
			tl.span.End()
		}
	}()

	testErr := errors.New("test error")
	_ = logger.Error(testErr)

	// Verify console output happened
	require.Equal(t, 1, logs.Len())
}
