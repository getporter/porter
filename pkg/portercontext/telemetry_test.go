package portercontext

import (
	"context"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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
	})
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
