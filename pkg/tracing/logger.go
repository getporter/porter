package tracing

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type RootLogger interface {
	StartSpan(ctx context.Context, attrs ...attribute.KeyValue) (context.Context, ScopedLogger)
	StartSpanWithName(ctx context.Context, op string, attrs ...attribute.KeyValue) (context.Context, ScopedLogger)
	Debug(span trace.Span, msg string, attrs ...attribute.KeyValue)
	Debugf(span trace.Span, format string, args ...interface{})
	Info(span trace.Span, msg string, attrs ...attribute.KeyValue)
	Infof(span trace.Span, format string, args ...interface{})
	Warn(span trace.Span, msg string, attrs ...attribute.KeyValue)
	Warnf(span trace.Span, format string, args ...interface{})
	Error(span trace.Span, err error, attrs ...attribute.KeyValue) error
	Errorf(span trace.Span, msg string, args ...interface{}) error
}

var _ RootLogger = traceLogger{}

type traceLogger struct {
	Logger *zap.Logger
	Tracer trace.Tracer

	// These are attributes that should be set on any root span
	rootAttrs []attribute.KeyValue
}

func NewLogger(logger *zap.Logger, tracer trace.Tracer, attrs ...attribute.KeyValue) RootLogger {
	if logger == nil {
		logger = zap.NewNop()
	}
	if tracer == nil {
		tracer = trace.NewNoopTracerProvider().Tracer("noop")
	}
	return traceLogger{
		Logger:    logger,
		Tracer:    tracer,
		rootAttrs: attrs,
	}
}

// StartSpanWithName creates a span with the name provided by the caller.
func (l traceLogger) StartSpanWithName(ctx context.Context, op string, attrs ...attribute.KeyValue) (context.Context, ScopedLogger) {
	type HasParentSpan interface {
		ParentSpanID() trace.SpanID
	}

	childCtx, childSpan := l.Tracer.Start(ctx, op)

	childSpan.SetAttributes(attrs...)

	isRootSpan := true
	if s, ok := childSpan.(HasParentSpan); ok {
		isRootSpan = !s.ParentSpanID().IsValid()
	}

	if isRootSpan {
		childSpan.SetAttributes(l.rootAttrs...)
	}

	return NewScopedLogger(childCtx, childSpan, l)
}

// StartSpanWithName creates a span with the caller's function name.
func (l traceLogger) StartSpan(ctx context.Context, attrs ...attribute.KeyValue) (context.Context, ScopedLogger) {
	return l.StartSpanWithName(ctx, callerFunc(0), attrs...)
}

func (l traceLogger) Debug(span trace.Span, msg string, attrs ...attribute.KeyValue) {
	l.Logger.Debug(msg, ConvertAttributesToFields(attrs)...)

	addLogToSpan(span, msg, "debug", attrs...)
}

func (l traceLogger) Debugf(span trace.Span, format string, args ...interface{}) {
	l.Debug(span, fmt.Sprintf(format, args...))
}

// Info records an informational log message and appends the specified key/value pairs.
func (l traceLogger) Info(span trace.Span, msg string, attrs ...attribute.KeyValue) {
	l.Logger.Info(msg, ConvertAttributesToFields(attrs)...)

	addLogToSpan(span, msg, "info", attrs...)
}

// Infof records an informational log message and formats it using the specified arguments.
func (l traceLogger) Infof(span trace.Span, format string, args ...interface{}) {
	l.Info(span, fmt.Sprintf(format, args...))
}

func addLogToSpan(span trace.Span, msg string, level string, attrs ...attribute.KeyValue) {
	attrs = append(attrs,
		attribute.String("level", level))
	span.AddEvent(msg, trace.WithAttributes(attrs...))
}

func (l traceLogger) Warn(span trace.Span, msg string, attrs ...attribute.KeyValue) {
	l.Logger.Warn(msg, ConvertAttributesToFields(attrs)...)

	addLogToSpan(span, msg, "warn", attrs...)
}

func (l traceLogger) Warnf(span trace.Span, format string, args ...interface{}) {
	l.Warn(span, fmt.Sprintf(format, args...))
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

// Error records the error and appends the specified key/value pairs.
// This function is a no-op, similar to errors.Wrap, when error is nil.
// This allows you to use it in your return statements and an error is only
// logged with it is not-nil.
func (l traceLogger) Error(span trace.Span, err error, attrs ...attribute.KeyValue) error {
	if err == nil {
		return err
	}

	msg := err.Error()
	l.Logger.Error(msg, ConvertAttributesToFields(attrs)...)

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

	span.RecordError(err, errOpts...)
	span.SetStatus(codes.Error, err.Error())

	return err
}

func (l traceLogger) Errorf(span trace.Span, msg string, args ...interface{}) error {
	return l.Error(span, errors.Errorf(msg, args...))
}

var _ RootLogger = consoleLogger{}

type consoleLogger struct {
	tracer trace.Tracer
}

// newConsoleLogger creates a logger that writes to the console (stderr)
// but does not attempt to trace.
// Do not use this unless we can't get our hands on the configured logger.
func newConsoleLogger() RootLogger {
	return consoleLogger{
		tracer: trace.NewNoopTracerProvider().Tracer("noop"),
	}
}

func (n consoleLogger) StartSpanWithName(ctx context.Context, op string, attrs ...attribute.KeyValue) (context.Context, ScopedLogger) {
	childCtx, childSpan := n.tracer.Start(ctx, op)
	childSpan.SetAttributes(attrs...)
	return NewScopedLogger(childCtx, childSpan, n)
}

func (n consoleLogger) StartSpan(ctx context.Context, attrs ...attribute.KeyValue) (context.Context, ScopedLogger) {
	return n.StartSpanWithName(ctx, callerFunc(0), attrs...)
}

func (n consoleLogger) StartSpanInLog(ctx context.Context, attrs ...attribute.KeyValue) (context.Context, ScopedLogger) {
	return n.StartSpanWithName(ctx, callerFunc(0), attrs...)
}

func (n consoleLogger) Debug(span trace.Span, msg string, attrs ...attribute.KeyValue) {
	fmt.Fprintln(os.Stderr, msg)
}

func (n consoleLogger) Debugf(span trace.Span, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func (n consoleLogger) Info(span trace.Span, msg string, attrs ...attribute.KeyValue) {
	fmt.Fprintln(os.Stderr, msg)
}

func (n consoleLogger) Infof(span trace.Span, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func (n consoleLogger) Warn(span trace.Span, msg string, attrs ...attribute.KeyValue) {
	fmt.Fprintln(os.Stderr, msg)
}

func (n consoleLogger) Warnf(span trace.Span, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func (n consoleLogger) Error(span trace.Span, err error, attrs ...attribute.KeyValue) error {
	fmt.Fprintln(os.Stderr, err)
	return err
}

func (n consoleLogger) Errorf(span trace.Span, msg string, args ...interface{}) error {
	return n.Error(span, errors.Errorf(msg, args...))
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
