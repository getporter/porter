package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestTraceLogger_ShouldLog(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	tracer := trace.NewNoopTracerProvider().Tracer("noop")
	l := newTraceLogger(context.Background(), nil, logger, NewTracer(tracer, nil))

	assert.True(t, l.ShouldLog(zap.ErrorLevel))
	assert.True(t, l.ShouldLog(zap.WarnLevel))
	assert.False(t, l.ShouldLog(zap.InfoLevel))
	assert.False(t, l.ShouldLog(zap.DebugLevel))
}
