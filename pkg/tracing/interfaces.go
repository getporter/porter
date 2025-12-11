package tracing

import (
	"go.opentelemetry.io/otel/attribute"
)

// ErrorTracer handles recording errors and events to OpenTelemetry spans.
// It NEVER outputs to console - only records to distributed tracing.
type ErrorTracer interface {
	// RecordError records an error to the current span without logging to console.
	// This should be used by all intermediate functions that handle/wrap errors.
	// Example: return tracer.RecordError(fmt.Errorf("could not load bundle: %w", err))
	RecordError(err error, attrs ...attribute.KeyValue) error

	// RecordErrorf is a convenience method that formats an error and records it.
	RecordErrorf(format string, args ...interface{}) error
}

// ErrorLogger handles logging errors to the console.
// This should ONLY be used at the top level (command handlers).
type ErrorLogger interface {
	// Error logs an error to the console AND records it to the span.
	// This should ONLY be called by top-level command handlers.
	// Example: return log.Error(err)
	Error(err error, attrs ...attribute.KeyValue) error

	// Errorf formats and logs an error to the console AND records it to the span.
	Errorf(format string, args ...interface{}) error
}
