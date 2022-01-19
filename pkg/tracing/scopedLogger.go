package tracing

import (
	"context"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ScopedLogger interface {
	StartSpan(op string, attrs ...attribute.KeyValue) (context.Context, ScopedLogger)
	StartSpanNamed(attrs ...attribute.KeyValue) (context.Context, ScopedLogger)
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

var _ ScopedLogger = scopedTraceLogger{}

type scopedTraceLogger struct {
	ctx        context.Context
	span       trace.Span
	rootLogger RootLogger
}

func LoggerFromContext(ctx context.Context) ScopedLogger {
	span := trace.SpanFromContext(ctx)

	if logger, ok := ctx.Value("porter.logger").(RootLogger); ok {
		_, l := NewScopedLogger(ctx, span, logger)
		return l
	}

	// The context we were passed didn't have a logger associated with it
	// This is a bit of a bug, but we can gracefully handle it by
	// printing directly to the console.
	_, l := NewScopedLogger(ctx, span, newConsoleLogger())
	return l
}

func NewScopedLogger(ctx context.Context, span trace.Span, logger RootLogger) (context.Context, ScopedLogger) {
	childCtx := context.WithValue(ctx, "porter.logger", logger)
	l := scopedTraceLogger{
		ctx:        childCtx,
		span:       span,
		rootLogger: logger,
	}
	return childCtx, l
}

func (l scopedTraceLogger) EndSpan(opts ...trace.SpanEndOption) {
	l.span.End(opts...)
}

func (l scopedTraceLogger) StartSpan(op string, attrs ...attribute.KeyValue) (context.Context, ScopedLogger) {
	return l.rootLogger.StartSpan(l.ctx, op, attrs...)
}

func (l scopedTraceLogger) StartSpanNamed(attrs ...attribute.KeyValue) (context.Context, ScopedLogger) {
	return l.rootLogger.StartSpan(l.ctx, callerFunc(0), attrs...)
}

func (l scopedTraceLogger) Debug(msg string, attrs ...attribute.KeyValue) {
	l.rootLogger.Debug(l.span, msg, attrs...)
}

func (l scopedTraceLogger) Debugf(format string, args ...interface{}) {
	l.rootLogger.Debugf(l.span, format, args...)
}

func (l scopedTraceLogger) Info(msg string, attrs ...attribute.KeyValue) {
	l.rootLogger.Info(l.span, msg, attrs...)
}

func (l scopedTraceLogger) Infof(format string, args ...interface{}) {
	l.rootLogger.Infof(l.span, format, args...)
}

func (l scopedTraceLogger) Warn(msg string, attrs ...attribute.KeyValue) {
	l.rootLogger.Warn(l.span, msg, attrs...)
}

func (l scopedTraceLogger) Warnf(format string, args ...interface{}) {
	l.rootLogger.Warnf(l.span, format, args...)
}

func (l scopedTraceLogger) Error(err error, attrs ...attribute.KeyValue) error {
	return l.rootLogger.Error(l.span, err, attrs...)
}

func (l scopedTraceLogger) Errorf(msg string, args ...interface{}) error {
	return l.rootLogger.Errorf(l.span, msg, args...)
}

func callerFunc(frames int) string {
	var pc [1]uintptr
	if runtime.Callers(frames+3, pc[:]) != 1 {
		return "unknown"
	}
	frame, _ := runtime.CallersFrames(pc[:]).Next()
	if frame.Function == "" {
		return "unknown"
	}
	slash_pieces := strings.Split(frame.Function, "/")
	dot_pieces := strings.SplitN(slash_pieces[len(slash_pieces)-1], ".", 2)
	return dot_pieces[len(dot_pieces)-1]
}
