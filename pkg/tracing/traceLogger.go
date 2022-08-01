package tracing

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TraceLogger how porter emits traces and logs to any configured listeners.
type TraceLogger interface {
	// StartSpan retrieves a logger from the current context and starts a new span
	// named after the current function.
	StartSpan(attrs ...attribute.KeyValue) (context.Context, TraceLogger)

	// StartSpanWithName retrieves a logger from the current context and starts a span with
	// the specified name.
	StartSpanWithName(ops string, attrs ...attribute.KeyValue) (context.Context, TraceLogger)

	// SetAttributes applies additional key/value pairs to the current trace span.
	SetAttributes(attrs ...attribute.KeyValue)

	// EndSpan finishes the span and submits it to the otel endpoint.
	EndSpan(opts ...trace.SpanEndOption)

	// Debug logs a message at the debug level.
	Debug(msg string, attrs ...attribute.KeyValue)

	// Debugf formats a message and logs it at the debug level.
	Debugf(format string, args ...interface{})

	// Info logs a message at the info level.
	Info(msg string, attrs ...attribute.KeyValue)

	// Infof formats a message and logs it at the info level.
	Infof(format string, args ...interface{})

	// Warn logs a message at the warning level.
	Warn(msg string, attrs ...attribute.KeyValue)

	// Warnf formats a message and prints it at the warning level.
	Warnf(format string, args ...interface{})

	// Error logs a message at the error level, when the specified error is not nil,
	// and marks the current span as failed.
	// Example: return log.Error(err)
	// Only log it in the function that generated the error, not when bubbling
	// it up the call stack.
	Error(err error, attrs ...attribute.KeyValue) error

	// Errorf logs a message at the error level and marks the current span as failed.
	Errorf(format string, arg ...interface{}) error

	// ShouldLog returns if the current log level includes the specified level.
	ShouldLog(level zapcore.Level) bool

	// IsTracingEnabled returns if the current logger is configed to send trace data.
	IsTracingEnabled() bool
}

type RootTraceLogger interface {
	TraceLogger

	// Close the tracer and send data to the telemetry collector
	Close()
}

var _ TraceLogger = traceLogger{}

// traceLogger writes tracing data to an open telemetry collector,
// and sends logs to both the tracer and a standard logger.
// This results in open telemetry collecting logs and traces,
// and also having all logs available in a file and send to the console.
type traceLogger struct {
	ctx    context.Context
	span   trace.Span
	logger *zap.Logger
	tracer Tracer
}

// Close the root span and send the telemetry data to the collector.
func (l traceLogger) Close() {
	l.span.End()

	if err := l.tracer.Close(context.Background()); err != nil {
		l.Errorf("error closing the Tracer: %w", err)
	}
}

// ShouldLog returns if the current log level includes the specified level.
func (l traceLogger) ShouldLog(level zapcore.Level) bool {
	return l.logger.Core().Enabled(level)
}

func (l traceLogger) IsTracingEnabled() bool {
	return !l.tracer.IsNoOp
}

// NewRootLogger creates a new TraceLogger and stores in on the context
func NewRootLogger(ctx context.Context, span trace.Span, logger *zap.Logger, tracer Tracer) (context.Context, RootTraceLogger) {
	childCtx := context.WithValue(ctx, contextKeyTraceLogger, traceLoggerContext{logger, tracer})
	return childCtx, newTraceLogger(childCtx, span, logger, tracer)
}

func newTraceLogger(ctx context.Context, span trace.Span, logger *zap.Logger, tracer Tracer) traceLogger {
	l := traceLogger{
		ctx:    ctx,
		span:   span,
		logger: logger,
		tracer: tracer,
	}
	return l
}

// EndSpan finishes the span and submits it to the otel endpoint.
func (l traceLogger) EndSpan(opts ...trace.SpanEndOption) {
	defer l.span.End(opts...)

	// If there was a panic, mark the span
	if p := recover(); p != nil {
		l.Errorf("panic: %s", p)
		panic(p) // retrow
	}
}

// StartSpan retrieves a logger from the current context and starts a new span
// named after the current function.
func (l traceLogger) StartSpan(attrs ...attribute.KeyValue) (context.Context, TraceLogger) {
	return l.StartSpanWithName(callerFunc(), attrs...)
}

// StartSpanWithName retrieves a logger from the current context and starts a span with
// the specified name.
func (l traceLogger) StartSpanWithName(op string, attrs ...attribute.KeyValue) (context.Context, TraceLogger) {
	childCtx, childSpan := l.tracer.Start(l.ctx, op)
	childSpan.SetAttributes(attrs...)
	return childCtx, newTraceLogger(childCtx, childSpan, l.logger, l.tracer)
}

// SetAttributes applies additional key/value pairs to the current trace span.
func (l traceLogger) SetAttributes(attrs ...attribute.KeyValue) {
	l.span.SetAttributes(attrs...)
}

// Debug logs a message at the debug level.
func (l traceLogger) Debug(msg string, attrs ...attribute.KeyValue) {
	l.logger.Debug(msg, convertAttributesToFields(attrs)...)

	addLogToSpan(l.span, msg, "debug", attrs...)
}

// Debugf formats a message and logs it at the debug level.
func (l traceLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Info logs a message at the info level.
func (l traceLogger) Info(msg string, attrs ...attribute.KeyValue) {
	l.logger.Info(msg, convertAttributesToFields(attrs)...)

	addLogToSpan(l.span, msg, "info", attrs...)
}

// Infof formats a message and logs it at the info level.
func (l traceLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warn logs a message at the warning level.
func (l traceLogger) Warn(msg string, attrs ...attribute.KeyValue) {
	l.logger.Warn(msg, convertAttributesToFields(attrs)...)

	addLogToSpan(l.span, msg, "warn", attrs...)
}

// Warnf formats a message and prints it at the warning level.
func (l traceLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l traceLogger) Errorf(format string, args ...interface{}) error {
	return l.Error(fmt.Errorf(format, args...))
}

// Error logs a message at the error level, when the specified error is not nil.
func (l traceLogger) Error(err error, attrs ...attribute.KeyValue) error {
	if err == nil {
		return err
	}

	msg := err.Error()
	l.logger.Error(msg, convertAttributesToFields(attrs)...)

	attrs = append(attrs, attribute.String("level", "error"))

	// Try to include the stack trace
	// I'm not using trace.WithStackTrace because it records the stack trace from _here_
	// and not the one attached to the error...
	errOpts := []trace.EventOption{
		trace.WithAttributes(attrs...),
	}

	errOpts = append(errOpts, trace.WithAttributes(
		semconv.ExceptionStacktraceKey.String(fmt.Sprintf("%+v", err)),
	))

	l.span.RecordError(err, errOpts...)
	l.span.SetStatus(codes.Error, err.Error())

	return err
}

// appends logs to a otel span as events
func addLogToSpan(span trace.Span, msg string, level string, attrs ...attribute.KeyValue) {
	attrs = append(attrs,
		attribute.String("level", level))
	span.AddEvent(msg, trace.WithAttributes(attrs...))
}

func callerFunc() string {
	callerUnknown := "unknown"

	// Depending on how this is called, the real function that we are targeting
	// may be a variable number of frames up the stack
	// Only look 10 frames up the stack
	for i := 0; i < 10; i++ {
		var pc [1]uintptr
		if runtime.Callers(i+2, pc[:]) != 1 {
			return callerUnknown
		}
		// translate the PC into function information
		frame, _ := runtime.CallersFrames(pc[:]).Next()
		if frame.Function == "" {
			return callerUnknown
		}

		// Locate the function that first called into the tracing package
		if strings.HasPrefix(frame.Function, "get.porter.sh/porter/pkg/tracing.") {
			continue
		}

		return strings.TrimPrefix(frame.Function, "get.porter.sh/porter/")
	}

	return callerUnknown
}
