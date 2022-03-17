package portercontext

import (
	"testing"

	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
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
