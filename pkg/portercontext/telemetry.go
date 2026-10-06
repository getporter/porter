package portercontext

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

const (
	// EnvSensitiveValues is the name of the environment variable used to pass
	// the sensitive values to a child porter process, e.g. a mixin, as a json
	// list of base64 encoded values, so that the child masks them in its trace data.
	EnvSensitiveValues = "PORTER_SENSITIVE_VALUES"

	// envTelemetryEnabled is the name of the environment variable that controls if trace data is exported.
	envTelemetryEnabled = "PORTER_TELEMETRY_ENABLED"

	// maxSensitiveValuesEnvSize is the largest value, in bytes, that we set
	// EnvSensitiveValues to. Linux limits a single environment variable to 128KiB.
	maxSensitiveValuesEnvSize = 100_000
)

// tracePropagator defines how the current span and baggage are passed between
// processes, using the W3C Trace Context and Baggage formats.
var tracePropagator = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})

// TraceEnvironNames returns the names of the environment variables used to
// pass the current span to a child process: TRACEPARENT, TRACESTATE and BAGGAGE.
func TraceEnvironNames() []string {
	fields := tracePropagator.Fields()
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = strings.ToUpper(field)
	}
	return names
}

// TraceEnviron returns the environment variables, e.g. TRACEPARENT, that
// should be set on a child process so that it can continue the trace of the
// span in the specified context. Returns an empty map when there is no span.
//
// All of the variables in TraceEnvironNames are returned, with an empty value
// when it doesn't apply to the span, so that a stale value that the child
// would otherwise inherit isn't combined with the span.
func TraceEnviron(ctx context.Context) map[string]string {
	env := make(map[string]string, 3)
	if !trace.SpanContextFromContext(ctx).IsValid() {
		return env
	}

	carrier := propagation.MapCarrier{}
	tracePropagator.Inject(ctx, carrier)
	for _, name := range TraceEnvironNames() {
		env[name] = carrier[strings.ToLower(name)]
	}
	return env
}

// extractTraceParent reads the environment variables set by TraceEnviron in
// the parent process, and returns a context containing the parent's span.
func (c *Context) extractTraceParent(ctx context.Context) context.Context {
	carrier := propagation.MapCarrier{}
	for _, field := range tracePropagator.Fields() {
		if v, ok := c.LookupEnv(strings.ToUpper(field)); ok {
			carrier[field] = v
		}
	}
	return tracePropagator.Extract(ctx, carrier)
}

// loadSensitiveValues masks the sensitive values passed to us with
// EnvSensitiveValues by the porter process that called us.
// The variable is removed, also from the environment of our process, so that
// it isn't passed on to the commands that we run.
//
// When the sensitive values can't be read, telemetry is turned off for us and
// the commands that we run instead, so that the values are not exported.
func (c *Context) loadSensitiveValues() {
	encoded, ok := c.LookupEnv(EnvSensitiveValues)
	if !ok {
		return
	}
	c.Unsetenv(EnvSensitiveValues)
	os.Unsetenv(EnvSensitiveValues)

	// Each value is base64 encoded, json decodes that into the original bytes
	var vals [][]byte
	if err := json.Unmarshal([]byte(encoded), &vals); err != nil {
		c.telemetryDisabled = true
		c.Setenv(envTelemetryEnabled, "false")
		return
	}

	sensitiveValues := make([]string, len(vals))
	for i, val := range vals {
		sensitiveValues[i] = string(val)
	}
	c.SetSensitiveValues(sensitiveValues)
}

// SensitiveValuesEnviron returns the environment variables, in the form
// KEY=VALUE, that should be set on a child porter process that is given
// sensitive values, e.g. a mixin, so that it masks them in its trace data.
// Returns nothing when trace data isn't exported or there are no sensitive values.
//
// When the sensitive values are too large to pass to the child, telemetry is
// turned off for the child instead, so that the values are not exported.
func (c *Context) SensitiveValuesEnviron() []string {
	if !c.tracerInitalized || c.censoredWriter == nil {
		return nil
	}

	sensitiveValues := c.censoredWriter.GetSensitiveValues()
	if len(sensitiveValues) == 0 {
		return nil
	}

	// Base64 encode each value, which is how json encodes bytes, because a
	// value isn't always valid UTF-8, e.g. the contents of a binary file.
	vals := make([][]byte, len(sensitiveValues))
	for i, val := range sensitiveValues {
		vals[i] = []byte(val)
	}

	encoded, err := json.Marshal(vals)
	if err != nil || len(encoded) > maxSensitiveValuesEnvSize {
		return []string{envTelemetryEnabled + "=false"}
	}
	return []string{EnvSensitiveValues + "=" + string(encoded)}
}

// censoredExporter masks sensitive values in the trace data before it is exported.
type censoredExporter struct {
	sdktrace.SpanExporter
	censoredWriter *CensoredWriter
}

func (e censoredExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if len(e.censoredWriter.values()) == 0 {
		return e.SpanExporter.ExportSpans(ctx, spans)
	}

	censored := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, span := range spans {
		censored[i] = censoredSpan{ReadOnlySpan: span, censoredWriter: e.censoredWriter}
	}
	return e.SpanExporter.ExportSpans(ctx, censored)
}

// censoredSpan is a span with the sensitive values masked in its name,
// status, attributes and events.
type censoredSpan struct {
	sdktrace.ReadOnlySpan
	censoredWriter *CensoredWriter
}

func (s censoredSpan) Name() string {
	return s.censoredWriter.CensorString(s.ReadOnlySpan.Name())
}

func (s censoredSpan) Status() sdktrace.Status {
	status := s.ReadOnlySpan.Status()
	status.Description = s.censoredWriter.CensorString(status.Description)
	return status
}

func (s censoredSpan) Attributes() []attribute.KeyValue {
	return s.censorAttributes(s.ReadOnlySpan.Attributes())
}

func (s censoredSpan) Events() []sdktrace.Event {
	events := s.ReadOnlySpan.Events()
	censored := make([]sdktrace.Event, len(events))
	for i, event := range events {
		event.Name = s.censoredWriter.CensorString(event.Name)
		event.Attributes = s.censorAttributes(event.Attributes)
		censored[i] = event
	}
	return censored
}

func (s censoredSpan) censorAttributes(attrs []attribute.KeyValue) []attribute.KeyValue {
	censored := make([]attribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		switch attr.Value.Type() {
		case attribute.STRING:
			attr.Value = attribute.StringValue(s.censoredWriter.CensorString(attr.Value.AsString()))
		case attribute.STRINGSLICE:
			vals := attr.Value.AsStringSlice()
			censoredVals := make([]string, len(vals))
			for j, val := range vals {
				censoredVals[j] = s.censoredWriter.CensorString(val)
			}
			attr.Value = attribute.StringSliceValue(censoredVals)
		}
		censored[i] = attr
	}
	return censored
}

func (c *Context) configureTelemetry(ctx context.Context, cfg LogConfiguration, logger *zap.Logger) error {
	if cfg.TelemetryServiceName == "" {
		cfg.TelemetryServiceName = "porter"
	}

	c.tracer = createNoopTracer()

	tracer, err := c.createTracer(ctx, cfg, logger)
	if err != nil {
		return err
	}

	// Only assign the tracer if one was configured (i.e. not noop)
	if !tracer.IsNoOp {
		c.tracer = tracer
		c.tracerInitalized = true
	}
	return nil
}

func createNoopTracer() tracing.Tracer {
	tracer := noop.NewTracerProvider().Tracer("noop")
	cleanup := func(_ context.Context) error { return nil }
	t := tracing.NewTracer(tracer, cleanup)
	t.IsNoOp = true
	return t
}

func (c *Context) createTracer(ctx context.Context, cfg LogConfiguration, logger *zap.Logger) (tracing.Tracer, error) {
	client, err := c.createTraceClient(c.logCfg)
	if err != nil {
		return tracing.Tracer{}, err
	}
	if client == nil {
		logger.Debug("telemetry disabled")
		return createNoopTracer(), nil
	}

	var exporter sdktrace.SpanExporter
	if cfg.TelemetryRedirectToFile {
		// Instead of sending trace data to a collector
		// save them to a file so that we can test our traces
		testTracePath := filepath.Join(cfg.TelemetryDirectory, c.buildLogFileName())
		logger.Debug(fmt.Sprintf("redirecting open telemetry trace data to a file: %s", testTracePath))

		tracesDir := filepath.Dir(testTracePath)
		if err = c.FileSystem.MkdirAll(tracesDir, pkg.FileModeDirectory); err != nil {
			return tracing.Tracer{}, fmt.Errorf("could not create traces directory at  %s: %w", tracesDir, err)
		}

		f, err := c.FileSystem.OpenFile(testTracePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, pkg.FileModeWritable)
		if err != nil {
			return tracing.Tracer{}, fmt.Errorf("could not create test traces file at  %s: %w", testTracePath, err)
		}
		c.traceFile = f
		exporter, err = stdouttrace.New(stdouttrace.WithWriter(f))
		if err != nil {
			return tracing.Tracer{}, fmt.Errorf("error creating a file trace exporter: %w", err)
		}
	} else {
		createTraceCtx, cancel := context.WithTimeout(ctx, cfg.TelemetryStartTimeout)
		defer cancel()
		exporter, err = otlptrace.New(createTraceCtx, client)
		if err != nil {
			return tracing.Tracer{}, fmt.Errorf("error creating an open telemetry trace exporter: %w", err)
		}
	}

	if c.censoredWriter != nil {
		exporter = censoredExporter{SpanExporter: exporter, censoredWriter: c.censoredWriter}
	}

	serviceVersion := pkg.Version
	if serviceVersion == "" {
		serviceVersion = "dev"
	}
	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.TelemetryServiceName),
		semconv.ServiceVersionKey.String(serviceVersion),
	)

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
	)
	otel.SetTextMapPropagator(tracePropagator)

	tracer := provider.Tracer("") // empty tracer name defaults to the underlying trace implementor
	cleanup := func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}
	return tracing.NewTracer(tracer, cleanup), nil
}

// createTraceClient from the Porter configuration
// See https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace
func (c *Context) createTraceClient(cfg LogConfiguration) (otlptrace.Client, error) {
	if !cfg.TelemetryEnabled {
		return nil, nil
	}

	switch cfg.TelemetryProtocol {
	case "grpc":
		opts := []otlptracegrpc.Option{}
		if cfg.TelemetryEndpoint != "" {
			opts = append(opts, otlptracegrpc.WithEndpoint(cfg.TelemetryEndpoint))
		}
		if cfg.TelemetryInsecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		if cfg.TelemetryCertificate != "" {
			creds, err := credentials.NewClientTLSFromFile(cfg.TelemetryCertificate, "")
			if err != nil {
				return nil, fmt.Errorf("invalid telemetry certificate in %s: %w", cfg.TelemetryCertificate, err)
			}
			opts = append(opts, otlptracegrpc.WithTLSCredentials(creds))
		}
		if cfg.TelemetryTimeout != "" {
			timeout, err := time.ParseDuration(cfg.TelemetryTimeout)
			if err != nil {
				return nil, fmt.Errorf("invalid telemetry timeout %s: %w", cfg.TelemetryTimeout, err)
			}
			opts = append(opts, otlptracegrpc.WithTimeout(timeout))
		}
		if cfg.TelemetryCompression != "" {
			opts = append(opts, otlptracegrpc.WithCompressor(cfg.TelemetryCompression))
		}
		if len(cfg.TelemetryHeaders) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(cfg.TelemetryHeaders))
		}
		return otlptracegrpc.NewClient(opts...), nil
	case "http/protobuf", "":
		var opts []otlptracehttp.Option
		if cfg.TelemetryEndpoint != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(cfg.TelemetryEndpoint))
		}
		if cfg.TelemetryInsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		if cfg.TelemetryCertificate != "" {
			certB, err := c.FileSystem.ReadFile(cfg.TelemetryCertificate)
			if err != nil {
				return nil, fmt.Errorf("invalid telemetry certificate in %s: %w", cfg.TelemetryCertificate, err)
			}
			cp := x509.NewCertPool()
			if ok := cp.AppendCertsFromPEM(certB); !ok {
				return nil, fmt.Errorf("could not use certificate in %s", cfg.TelemetryCertificate)
			}
			opts = append(opts, otlptracehttp.WithTLSClientConfig(&tls.Config{RootCAs: cp}))
		}
		if cfg.TelemetryTimeout != "" {
			timeout, err := time.ParseDuration(cfg.TelemetryTimeout)
			if err != nil {
				return nil, fmt.Errorf("invalid telemetry timeout %s. Supported values are durations such as 30s or 1m: %w", cfg.TelemetryTimeout, err)
			}
			opts = append(opts, otlptracehttp.WithTimeout(timeout))
		}
		if cfg.TelemetryCompression != "" {
			var compression otlptracehttp.Compression
			switch cfg.TelemetryCompression {
			case "gzip":
				compression = otlptracehttp.GzipCompression
			default:
				compression = otlptracehttp.NoCompression
			}
			opts = append(opts, otlptracehttp.WithCompression(compression))
		}
		if len(cfg.TelemetryHeaders) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(cfg.TelemetryHeaders))
		}
		return otlptracehttp.NewClient(opts...), nil
	default:
		return nil, fmt.Errorf("invalid OTEL_EXPORTER_OTLP_PROTOCOL value %s. Only grpc and http/protobuf are supported", cfg.TelemetryProtocol)
	}
}
