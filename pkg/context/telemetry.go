package context

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"time"

	"get.porter.sh/porter/pkg"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func FromContext(ctx context.Context) (*Context, bool) {
	val := ctx.Value("porter.context")
	pc, ok := val.(*Context)
	return pc, ok
}

func (c *Context) configureTelemetry(logger *zap.Logger, cfg LogConfiguration) error {
	// default to noop
	c.tracer = trace.NewNoopTracerProvider().Tracer("noop")
	c.traceCloser = nil

	client, err := c.createTraceClient(cfg)
	if err != nil {
		return err
	}
	if client == nil {
		logger.Debug("telemetry disabled")
		return nil
	}

	exporter, err := otlptrace.New(context.TODO(), client)
	if err != nil {
		log.Fatalf("creating OTLP trace exporter: %v", err)
	}

	if c.traceServiceName == "" {
		c.traceServiceName = "porter"
	}
	serviceVersion := pkg.Version
	if serviceVersion == "" {
		serviceVersion = "dev"
	}
	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(c.traceServiceName),
		semconv.ServiceVersionKey.String(serviceVersion),
	)

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
	)

	c.tracer = provider.Tracer("") // empty tracer name defaults to the underlying trace implementor
	c.traceCloser = provider
	return nil
}

// createTraceClient from the Porter configuration
// See https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace
func (c *Context) createTraceClient(cfg LogConfiguration) (otlptrace.Client, error) {
	if !cfg.TelemetryEnabled {
		return nil, nil
	}

	switch cfg.TelemetryProtocol {
	case "grpc":
		opts := []otlptracegrpc.Option{otlptracegrpc.WithDialOption(grpc.WithBlock())}
		if cfg.TelemetryEndpoint != "" {
			opts = append(opts, otlptracegrpc.WithEndpoint(cfg.TelemetryEndpoint))
		}
		if cfg.TelemetryInsecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		if cfg.TelemetryCertificate != "" {
			creds, err := credentials.NewClientTLSFromFile(cfg.TelemetryCertificate, "")
			if err != nil {
				return nil, errors.Wrapf(err, "invalid telemetry certificate in %s", cfg.TelemetryCertificate)
			}
			opts = append(opts, otlptracegrpc.WithTLSCredentials(creds))
		}
		if cfg.TelemetryTimeout != "" {
			timeout, err := time.ParseDuration(cfg.TelemetryTimeout)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid telemetry timeout %s", cfg.TelemetryTimeout)
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
				return nil, errors.Wrapf(err, "invalid telemetry certificate in %s", cfg.TelemetryCertificate)
			}
			cp := x509.NewCertPool()
			if ok := cp.AppendCertsFromPEM(certB); !ok {
				return nil, errors.Errorf("could not use certificate in %s", cfg.TelemetryCertificate)
			}
			opts = append(opts, otlptracehttp.WithTLSClientConfig(&tls.Config{RootCAs: cp}))
		}
		if cfg.TelemetryTimeout != "" {
			timeout, err := time.ParseDuration(cfg.TelemetryTimeout)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid telemetry timeout %s. Supported values are durations such as 30s or 1m.", cfg.TelemetryTimeout)
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
		return nil, errors.Errorf("invalid OTEL_EXPORTER_OTLP_PROTOCOL value %s. Only grpc and http/protobuf are supported", cfg.TelemetryProtocol)
	}
}
