package tracing

import "context"

// GetTracer retrieves the ErrorTracer from the context.
// This is a convenience function for accessing just the tracer.
func GetTracer(ctx context.Context) ErrorTracer {
	return LoggerFromContext(ctx)
}

// GetLogger retrieves the ErrorLogger from the context.
// This is a convenience function for accessing just the logger.
func GetLogger(ctx context.Context) TraceLogger {
	return LoggerFromContext(ctx)
}
