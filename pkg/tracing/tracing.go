package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// custom type for looking up values from context.Context
type contextKey string

// key for retrieving the TraceLogger stored on the context
const contextKeyTraceLogger = contextKey("porter.traceLogger")

// stores data in context.Context to recreate a TraceLogger
type traceLoggerContext struct {
	logger *zap.Logger
	tracer trace.Tracer
}

// LoggerFromContext retrieves a logger from the specified context.
// When the context is missing a logger/tracer, no-op implementations are provided.
func LoggerFromContext(ctx context.Context) traceLogger {
	span := trace.SpanFromContext(ctx)

	var logger *zap.Logger
	var tracer trace.Tracer
	if tl, ok := ctx.Value(contextKeyTraceLogger).(traceLoggerContext); ok {
		logger = tl.logger
		tracer = tl.tracer
	} else {
		// default to no-op
		logger = zap.NewNop()
		tracer = trace.NewNoopTracerProvider().Tracer("noop")
	}

	return newTraceLogger(ctx, span, logger, NewTracer(tracer, nil))
}

// StartSpan retrieves a logger from the current context and starts a new span
// named after the current function.
func StartSpan(ctx context.Context, attrs ...attribute.KeyValue) (context.Context, TraceLogger) {
	log := LoggerFromContext(ctx)
	return log.StartSpan(attrs...)
}

func StartSpanForComponent(ctx context.Context, tracer Tracer, attrs ...attribute.KeyValue) (context.Context, TraceLogger) {
	log := LoggerFromContext(ctx)
	ctx, span := tracer.Start(ctx, callerFunc())

	l := newTraceLogger(ctx, span, log.logger, tracer)
	return ctx, l
}

// StartSpanWithName retrieves a logger from the current context and starts a span with
// the specified name.
func StartSpanWithName(ctx context.Context, op string, attrs ...attribute.KeyValue) (context.Context, TraceLogger) {
	log := LoggerFromContext(ctx)
	return log.StartSpanWithName(op, attrs...)
}

func convertAttributesToFields(attrs []attribute.KeyValue) []zap.Field {
	fields := make([]zap.Field, len(attrs))
	for i, attr := range attrs {
		fields[i] = convertAttributeToField(attr)
	}
	return fields
}

func convertAttributeToField(attr attribute.KeyValue) zapcore.Field {
	key := string(attr.Key)

	switch attr.Value.Type() {

	case attribute.BOOL:
		return zap.Bool(key, attr.Value.AsBool())
	case attribute.BOOLSLICE:
		return zap.Bools(key, attr.Value.AsBoolSlice())
	case attribute.FLOAT64:
		return zap.Float64(key, attr.Value.AsFloat64())
	case attribute.FLOAT64SLICE:
		return zap.Float64s(key, attr.Value.AsFloat64Slice())
	case attribute.INT64:
		return zap.Int64(key, attr.Value.AsInt64())
	case attribute.INT64SLICE:
		return zap.Int64s(key, attr.Value.AsInt64Slice())
	case attribute.STRING:
		return zap.String(key, attr.Value.AsString())
	case attribute.STRINGSLICE:
		return zap.Strings(key, attr.Value.AsStringSlice())
	default:
		// Give up and do our best
		b, _ := attr.Value.MarshalJSON()
		return zap.String(key, string(b))
	}
}
