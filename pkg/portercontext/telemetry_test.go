package portercontext

import (
	"context"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestContext_createTraceClient(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		c := NewTestContext(t)
		cfg := LogConfiguration{
			TelemetryEnabled: false,
		}
		client, err := c.createTraceClient(cfg)
		require.NoError(t, err)
		assert.Nil(t, client, "telemetry should be disabled")
	})

	t.Run("grpc", func(t *testing.T) {
		c := NewTestContext(t)
		c.UseFilesystem()
		cfg := LogConfiguration{
			TelemetryEnabled:     true,
			TelemetryEndpoint:    "192.168.0.1:13789",
			TelemetryProtocol:    "grpc",
			TelemetryInsecure:    true,
			TelemetryCertificate: "testdata/test-ca.pem",
			TelemetryCompression: "gzip",
			TelemetryTimeout:     "3s",
		}
		client, err := c.createTraceClient(cfg)
		require.NoError(t, err)
		require.NotNil(t, client, "expected a client to be returned")
		assert.IsType(t, otlptracegrpc.NewClient(), client, "expected a grpc client")
	})

	t.Run("grpc, invalid cert", func(t *testing.T) {
		c := NewTestContext(t)
		c.UseFilesystem()
		cfg := LogConfiguration{
			TelemetryEnabled:     true,
			TelemetryProtocol:    "grpc",
			TelemetryCertificate: "missingcert.pem",
		}
		_, err := c.createTraceClient(cfg)
		tests.RequireErrorContains(t, err, "invalid telemetry certificate")
	})

	t.Run("grpc, invalid timeout", func(t *testing.T) {
		c := NewTestContext(t)
		c.UseFilesystem()
		cfg := LogConfiguration{
			TelemetryEnabled:  true,
			TelemetryProtocol: "grpc",
			TelemetryTimeout:  "300",
		}
		_, err := c.createTraceClient(cfg)
		tests.RequireErrorContains(t, err, "invalid telemetry timeout")
	})

	t.Run("grpc, invalid compression defaults to none", func(t *testing.T) {
		c := NewTestContext(t)
		c.UseFilesystem()
		cfg := LogConfiguration{
			TelemetryEnabled:     true,
			TelemetryProtocol:    "grpc",
			TelemetryCompression: "oops",
		}
		_, err := c.createTraceClient(cfg)
		require.NoError(t, err)
	})

	t.Run("http", func(t *testing.T) {
		c := NewTestContext(t)
		c.UseFilesystem()
		cfg := LogConfiguration{
			TelemetryEnabled:     true,
			TelemetryEndpoint:    "192.168.0.1:13789",
			TelemetryProtocol:    "http/protobuf",
			TelemetryInsecure:    true,
			TelemetryCertificate: "testdata/test-ca.pem",
			TelemetryCompression: "gzip",
			TelemetryTimeout:     "3s",
		}
		client, err := c.createTraceClient(cfg)
		require.NoError(t, err)
		require.NotNil(t, client, "expected a client to be returned")
		assert.IsType(t, otlptracehttp.NewClient(), client, "expected a http client")
	})

	t.Run("http, invalid cert", func(t *testing.T) {
		c := NewTestContext(t)
		c.UseFilesystem()
		cfg := LogConfiguration{
			TelemetryEnabled:     true,
			TelemetryProtocol:    "http/protobuf",
			TelemetryCertificate: "missingcert.pem",
		}
		_, err := c.createTraceClient(cfg)
		tests.RequireErrorContains(t, err, "invalid telemetry certificate")
	})

	t.Run("grpc, invalid timeout", func(t *testing.T) {
		c := NewTestContext(t)
		c.UseFilesystem()
		cfg := LogConfiguration{
			TelemetryEnabled:  true,
			TelemetryProtocol: "http/protobuf",
			TelemetryTimeout:  "300",
		}
		_, err := c.createTraceClient(cfg)
		tests.RequireErrorContains(t, err, "invalid telemetry timeout")
	})

	t.Run("http, invalid compression defaults to none", func(t *testing.T) {
		c := NewTestContext(t)
		c.UseFilesystem()
		cfg := LogConfiguration{
			TelemetryEnabled:     true,
			TelemetryProtocol:    "http/protobuf",
			TelemetryCompression: "oops",
		}
		_, err := c.createTraceClient(cfg)
		require.NoError(t, err)
	})

	t.Run("invalid protocol", func(t *testing.T) {
		c := NewTestContext(t)
		cfg := LogConfiguration{
			TelemetryEnabled:  true,
			TelemetryProtocol: "oops",
		}
		_, err := c.createTraceClient(cfg)
		tests.RequireErrorContains(t, err, "invalid OTEL_EXPORTER_OTLP_PROTOCOL value")
	})
}

// useTestTracer configures the context with a tracer that creates real spans without exporting them.
func useTestTracer(c *TestContext) {
	provider := sdktrace.NewTracerProvider()
	c.tracer = tracing.NewTracer(provider.Tracer(""), provider.Shutdown)
}

func TestTraceEnviron(t *testing.T) {
	t.Run("no span", func(t *testing.T) {
		assert.Empty(t, TraceEnviron(context.Background()))
	})

	t.Run("span", func(t *testing.T) {
		c := NewTestContext(t)
		useTestTracer(c)

		ctx, log := c.StartRootSpan(context.Background(), t.Name())
		defer log.Close()

		env := TraceEnviron(ctx)
		sc := trace.SpanContextFromContext(ctx)
		require.Contains(t, env, "TRACEPARENT")
		assert.Contains(t, env["TRACEPARENT"], sc.TraceID().String())
		assert.Contains(t, env["TRACEPARENT"], sc.SpanID().String())

		// Always set, so that the child doesn't use a stale value that it inherited
		require.Contains(t, env, "TRACESTATE")
		assert.Empty(t, env["TRACESTATE"])
		require.Contains(t, env, "BAGGAGE")
		assert.Empty(t, env["BAGGAGE"])
	})

	t.Run("span with baggage", func(t *testing.T) {
		c := NewTestContext(t)
		useTestTracer(c)

		member, err := baggage.NewMember("owner", "me")
		require.NoError(t, err)
		bags, err := baggage.New(member)
		require.NoError(t, err)

		ctx, log := c.StartRootSpan(baggage.ContextWithBaggage(context.Background(), bags), t.Name())
		defer log.Close()

		assert.Equal(t, "owner=me", TraceEnviron(ctx)["BAGGAGE"])
	})
}

func TestTraceEnvironNames(t *testing.T) {
	assert.ElementsMatch(t, []string{"TRACEPARENT", "TRACESTATE", "BAGGAGE"}, TraceEnvironNames())
}

func TestContext_StartRootSpan_ContinuesParentTrace(t *testing.T) {
	// Start a span in the "parent process"
	parent := NewTestContext(t)
	useTestTracer(parent)
	parentCtx, parentLog := parent.StartRootSpan(context.Background(), "parent")
	defer parentLog.Close()
	parentSpan := trace.SpanContextFromContext(parentCtx)

	t.Run("parent in environment", func(t *testing.T) {
		child := NewTestContext(t)
		useTestTracer(child)
		for k, v := range TraceEnviron(parentCtx) {
			child.Setenv(k, v)
		}

		ctx, log := child.StartRootSpan(context.Background(), "child")
		defer log.Close()

		childSpan := trace.SpanContextFromContext(ctx)
		assert.Equal(t, parentSpan.TraceID(), childSpan.TraceID(), "the child should continue the parent's trace")
		assert.NotEqual(t, parentSpan.SpanID(), childSpan.SpanID(), "the child should have its own span")
	})

	t.Run("no parent", func(t *testing.T) {
		child := NewTestContext(t)
		useTestTracer(child)

		ctx, log := child.StartRootSpan(context.Background(), "child")
		defer log.Close()

		childSpan := trace.SpanContextFromContext(ctx)
		assert.True(t, childSpan.IsValid())
		assert.NotEqual(t, parentSpan.TraceID(), childSpan.TraceID(), "the child should start a new trace")
	})

	t.Run("invalid parent", func(t *testing.T) {
		child := NewTestContext(t)
		useTestTracer(child)
		child.Setenv("TRACEPARENT", "not-a-trace")

		ctx, log := child.StartRootSpan(context.Background(), "child")
		defer log.Close()

		assert.True(t, trace.SpanContextFromContext(ctx).IsValid(), "the child should start a new trace")
	})
}

func TestContext_CommandContext_PassesTrace(t *testing.T) {
	const inherited = "TRACEPARENT=00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"

	t.Run("span", func(t *testing.T) {
		c := NewTestContext(t)
		useTestTracer(c)
		c.Setenv("TRACEPARENT", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")

		ctx, log := c.StartRootSpan(context.Background(), t.Name())
		defer log.Close()

		cmd := c.CommandContext(ctx, "echo")
		want := "TRACEPARENT=" + TraceEnviron(ctx)["TRACEPARENT"]
		require.NotEqual(t, inherited, want)
		// When a variable is repeated, the last value is used
		assert.Equal(t, want, lastEnv(cmd.Env, "TRACEPARENT"), "the command should be a child of the current span")
	})

	t.Run("span, stale state in environment", func(t *testing.T) {
		c := NewTestContext(t)
		useTestTracer(c)

		// Start the span before the stale values are set, so that they aren't part of the span
		ctx, log := c.StartRootSpan(context.Background(), t.Name())
		defer log.Close()
		c.Setenv("TRACESTATE", "vendor=stale")
		c.Setenv("BAGGAGE", "owner=stale")

		cmd := c.CommandContext(ctx, "echo")
		assert.Equal(t, "TRACESTATE=", lastEnv(cmd.Env, "TRACESTATE"), "stale trace state should not be passed with the current span")
		assert.Equal(t, "BAGGAGE=", lastEnv(cmd.Env, "BAGGAGE"), "stale baggage should not be passed with the current span")
	})

	t.Run("no span", func(t *testing.T) {
		c := NewTestContext(t)
		c.Setenv("TRACEPARENT", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")

		cmd := c.CommandContext(context.Background(), "echo")
		assert.Equal(t, inherited, lastEnv(cmd.Env, "TRACEPARENT"), "the inherited value should be left alone")
	})
}

func lastEnv(environ []string, key string) string {
	var found string
	for _, env := range environ {
		if strings.HasPrefix(env, key+"=") {
			found = env
		}
	}
	return found
}

func TestCensoredExporter(t *testing.T) {
	stubs := tracetest.SpanStubs{
		{
			Name: "run topsecret",
			Attributes: []attribute.KeyValue{
				attribute.String("stdin", "arguments: [open_door, topsecret]"),
				attribute.StringSlice("args", []string{"open_door", "topsecret"}),
				attribute.Int("count", 1),
			},
			Events: []sdktrace.Event{
				{Name: "fail open_door topsecret", Attributes: []attribute.KeyValue{attribute.String("exception.message", "couldn't run fail open_door topsecret")}},
			},
			Status: sdktrace.Status{Code: codes.Error, Description: "couldn't run fail open_door topsecret"},
		},
	}

	t.Run("no sensitive values", func(t *testing.T) {
		inner := tracetest.NewInMemoryExporter()
		exporter := censoredExporter{SpanExporter: inner, censoredWriter: NewCensoredWriter(nil)}

		require.NoError(t, exporter.ExportSpans(context.Background(), stubs.Snapshots()))

		assert.Equal(t, stubs, inner.GetSpans(), "expected the spans to be exported unchanged")
	})

	t.Run("sensitive values", func(t *testing.T) {
		inner := tracetest.NewInMemoryExporter()
		censoredWriter := NewCensoredWriter(nil)
		censoredWriter.SetSensitiveValues([]string{"topsecret", " "})
		exporter := censoredExporter{SpanExporter: inner, censoredWriter: censoredWriter}

		require.NoError(t, exporter.ExportSpans(context.Background(), stubs.Snapshots()))

		got := inner.GetSpans()
		require.Len(t, got, 1)
		assert.Equal(t, "run *******", got[0].Name)
		assert.Equal(t, []attribute.KeyValue{
			attribute.String("stdin", "arguments: [open_door, *******]"),
			attribute.StringSlice("args", []string{"open_door", "*******"}),
			attribute.Int("count", 1),
		}, got[0].Attributes)
		require.Len(t, got[0].Events, 1)
		assert.Equal(t, "fail open_door *******", got[0].Events[0].Name)
		assert.Equal(t, []attribute.KeyValue{attribute.String("exception.message", "couldn't run fail open_door *******")}, got[0].Events[0].Attributes)
		assert.Equal(t, sdktrace.Status{Code: codes.Error, Description: "couldn't run fail open_door *******"}, got[0].Status)

		// The original spans must not be modified
		assert.Equal(t, "run topsecret", stubs[0].Name)
		assert.Equal(t, "arguments: [open_door, topsecret]", stubs[0].Attributes[0].Value.AsString())
	})
}

func TestContext_loadSensitiveValues(t *testing.T) {
	t.Run("not set", func(t *testing.T) {
		c := NewTestContext(t)
		c.Unsetenv(EnvSensitiveValues)

		c.loadSensitiveValues()

		assert.Empty(t, c.censoredWriter.GetSensitiveValues())
	})

	t.Run("set", func(t *testing.T) {
		c := NewTestContext(t)
		c.Setenv(EnvSensitiveValues, `["topsecret","multi\nline"]`)

		c.loadSensitiveValues()

		assert.Equal(t, []string{"topsecret", "multi\nline"}, c.censoredWriter.GetSensitiveValues())
		_, ok := c.LookupEnv(EnvSensitiveValues)
		assert.False(t, ok, "expected the sensitive values to not be passed on to the commands that we run")
	})

	t.Run("invalid", func(t *testing.T) {
		c := NewTestContext(t)
		c.Setenv(EnvSensitiveValues, `topsecret`)

		c.loadSensitiveValues()

		assert.Empty(t, c.censoredWriter.GetSensitiveValues())
		_, ok := c.LookupEnv(EnvSensitiveValues)
		assert.False(t, ok, "expected the sensitive values to not be passed on to the commands that we run")
	})
}

func TestContext_SensitiveValuesEnviron(t *testing.T) {
	t.Run("telemetry disabled", func(t *testing.T) {
		c := NewTestContext(t)
		c.SetSensitiveValues([]string{"topsecret"})

		assert.Empty(t, c.SensitiveValuesEnviron())
	})

	t.Run("no sensitive values", func(t *testing.T) {
		c := NewTestContext(t)
		c.tracerInitalized = true

		assert.Empty(t, c.SensitiveValuesEnviron())
	})

	t.Run("sensitive values", func(t *testing.T) {
		c := NewTestContext(t)
		c.tracerInitalized = true
		c.SetSensitiveValues([]string{"topsecret", "multi\nline"})

		assert.Equal(t, []string{`PORTER_SENSITIVE_VALUES=["topsecret","multi\nline"]`}, c.SensitiveValuesEnviron())
	})

	t.Run("too large", func(t *testing.T) {
		c := NewTestContext(t)
		c.tracerInitalized = true
		c.SetSensitiveValues([]string{strings.Repeat("a", maxSensitiveValuesEnvSize)})

		assert.Equal(t, []string{"PORTER_TELEMETRY_ENABLED=false"}, c.SensitiveValuesEnviron(),
			"expected telemetry to be disabled for the child when the sensitive values can't be passed to it")
	})
}
