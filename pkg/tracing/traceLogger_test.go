package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestExtractFuncName(t *testing.T) {
	for _, test := range []struct {
		input    string
		expected string
		ok       bool
	}{
		{"", "", false},
		{"porter/", "", false},
		{"porter/v.", "", false},
		{"github.com/getporter/porter/tracing.StartSpan", "StartSpan", true},
	} {
		fn, ok := extractFuncName(test.input)
		if fn != test.expected || ok != test.ok {
			t.Errorf("failed %q, got %q %v, expected %q %v", test.input, fn, ok, test.expected, test.ok)
		}
	}
}

func TestTraceLogger_ShouldLog(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	tracer := trace.NewNoopTracerProvider().Tracer("noop")
	l := newTraceLogger(context.Background(), nil, logger, tracer)

	assert.True(t, l.ShouldLog(zap.ErrorLevel))
	assert.True(t, l.ShouldLog(zap.WarnLevel))
	assert.False(t, l.ShouldLog(zap.InfoLevel))
	assert.False(t, l.ShouldLog(zap.DebugLevel))
}
