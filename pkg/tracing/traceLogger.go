package tracing

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TraceLogger how porter emits traces and logs to any configured listeners.
type TraceLogger interface {
	StartSpan(attrs ...attribute.KeyValue) (context.Context, TraceLogger)
	StartSpanWithName(ops string, attrs ...attribute.KeyValue) (context.Context, TraceLogger)
	EndSpan(opts ...trace.SpanEndOption)
	Debug(msg string, attrs ...attribute.KeyValue)
	Debugf(format string, args ...interface{})
	Info(msg string, attrs ...attribute.KeyValue)
	Infof(format string, args ...interface{})
	Warn(msg string, attrs ...attribute.KeyValue)
	Warnf(format string, args ...interface{})
	Error(err error, attrs ...attribute.KeyValue) error
	Errorf(msg string, args ...interface{}) error
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
	tracer trace.Tracer
}

// NewRootLogger creates a new TraceLogger and stores in on the context
func NewRootLogger(ctx context.Context, span trace.Span, logger *zap.Logger, tracer trace.Tracer) (context.Context, TraceLogger) {
	childCtx := context.WithValue(ctx, contextKeyTraceLogger, traceLoggerContext{logger, tracer})
	return childCtx, newTraceLogger(childCtx, span, logger, tracer)
}

func newTraceLogger(ctx context.Context, span trace.Span, logger *zap.Logger, tracer trace.Tracer) TraceLogger {
	l := traceLogger{
		ctx:    ctx,
		span:   span,
		logger: logger,
		tracer: tracer,
	}
	return l
}

func (l traceLogger) EndSpan(opts ...trace.SpanEndOption) {
	l.span.End(opts...)
}

func (l traceLogger) StartSpan(attrs ...attribute.KeyValue) (context.Context, TraceLogger) {
	return l.StartSpanWithName(callerFunc(0), attrs...)
}

func (l traceLogger) StartSpanWithName(op string, attrs ...attribute.KeyValue) (context.Context, TraceLogger) {
	childCtx, childSpan := l.tracer.Start(l.ctx, op)
	childSpan.SetAttributes(attrs...)
	return childCtx, newTraceLogger(childCtx, childSpan, l.logger, l.tracer)
}

func (l traceLogger) Debug(msg string, attrs ...attribute.KeyValue) {
	l.logger.Debug(msg, ConvertAttributesToFields(attrs)...)

	addLogToSpan(l.span, msg, "debug", attrs...)
}

func (l traceLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l traceLogger) Info(msg string, attrs ...attribute.KeyValue) {
	l.logger.Info(msg, ConvertAttributesToFields(attrs)...)

	addLogToSpan(l.span, msg, "info", attrs...)
}

func (l traceLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l traceLogger) Warn(msg string, attrs ...attribute.KeyValue) {
	l.logger.Warn(msg, ConvertAttributesToFields(attrs)...)

	addLogToSpan(l.span, msg, "warn", attrs...)
}

func (l traceLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l traceLogger) Error(err error, attrs ...attribute.KeyValue) error {
	if err == nil {
		return err
	}

	msg := err.Error()
	l.logger.Error(msg, ConvertAttributesToFields(attrs)...)

	attrs = append(attrs, attribute.String("level", "error"))

	// Try to include the stack trace
	// I'm not using trace.WithStackTrace because it records the stack trace from _here_
	// and not the one attached to the error...
	errOpts := []trace.EventOption{
		trace.WithAttributes(attrs...),
	}

	if st := findStack(err); st != nil {
		errOpts = append(errOpts, trace.WithAttributes(
			semconv.ExceptionStacktraceKey.String(fmt.Sprintf("%+v", st)),
		))
	}

	l.span.RecordError(err, errOpts...)
	l.span.SetStatus(codes.Error, err.Error())

	return err
}

func (l traceLogger) Errorf(msg string, args ...interface{}) error {
	return l.Error(errors.Errorf(msg, args...))
}

// appends logs to a otel span as events
func addLogToSpan(span trace.Span, msg string, level string, attrs ...attribute.KeyValue) {
	attrs = append(attrs,
		attribute.String("level", level))
	span.AddEvent(msg, trace.WithAttributes(attrs...))
}

// Find the closest stack trace to where the error was generated
// https://github.com/pkg/errors/issues/173#issuecomment-456729811
func findStack(err error) errors.StackTrace {
	type causer interface {
		Cause() error
	}

	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	var stackErr error

	for {
		if _, ok := err.(stackTracer); ok {
			stackErr = err
		}

		if causer, ok := err.(causer); ok {
			err = causer.Cause()
		} else {
			break
		}
	}

	if stackErr != nil {
		return stackErr.(stackTracer).StackTrace()
	}

	return nil
}

func callerFunc(frames int) string {
	var pc [1]uintptr
	// we expect there're 3 functions(runtime.Callers, callerFunc,
	// StartSpanNamedFromCaller) the stack before the actual caller functions
	// we would like to find. If no function is found on the stack, return
	// unknown
	if runtime.Callers(frames+3, pc[:]) != 1 {
		return "unknown"
	}
	// translate the PC into function information
	frame, _ := runtime.CallersFrames(pc[:]).Next()
	if frame.Function == "" {
		return "unknown"
	}

	fnName, ok := extractFuncName(frame.Function)
	if !ok {
		return "unknown"
	}

	return fnName

}

// extractFuncName returns function names from a qualified full import path.
// for example: "github.com/getporter/porter.ListInstallations", "main.Install"
func extractFuncName(fn string) (string, bool) {
	lastSlashIdx := strings.LastIndex(fn, "/")
	if lastSlashIdx+1 >= len(fn) {
		// a function name ended with a "/"
		return "", false
	}

	qualifiedName := fn[lastSlashIdx+1:]
	packageDotPos := strings.Index(qualifiedName, ".")
	if packageDotPos < 0 || packageDotPos+1 >= len(qualifiedName) {
		// qualifiedName ended with a "."
		return "", false
	}

	return qualifiedName[packageDotPos+1:], true
}
