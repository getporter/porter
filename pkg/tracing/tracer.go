package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type TracerCleanup func(ctx context.Context) error

// Tracer wraps an open telemetry tracer connection in Porter
// so that its cleanup function can be moved around with it.
type Tracer struct {
	trace.Tracer
	cleanup TracerCleanup
}

// NewTracer wraps an open telemetry tracer and its cleanup function.
func NewTracer(t trace.Tracer, cleanup TracerCleanup) Tracer {
	return Tracer{
		Tracer:  t,
		cleanup: cleanup,
	}
}

// Close the tracer, releasing resources and sending the data to the collector.
func (t Tracer) Close(ctx context.Context) error {
	if t.cleanup != nil {
		return t.cleanup(ctx)
	}
	return nil
}
