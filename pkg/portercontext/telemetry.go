package portercontext

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"time"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (c *Context) configureTelemetry(ctx context.Context, serviceName string, logger *zap.Logger) error {
	if serviceName == "" {
		serviceName = "porter"
	}

	c.tracer = createNoopTracer()

	tracer, err := c.createTracer(ctx, serviceName, logger)
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
	tracer := trace.NewNoopTracerProvider().Tracer("noop")
	cleanup := func(_ context.Context) error { return nil }
	t := tracing.NewTracer(tracer, cleanup)
	t.IsNoOp = true
	return t
}

func (c *Context) createTracer(ctx context.Context, serviceName string, logger *zap.Logger) (tracing.Tracer, error) {
	client, err := c.createTraceClient(c.logCfg)
	if err != nil {
		return tracing.Tracer{}, err
	}
	if client == nil {
		logger.Debug("telemetry disabled")
		return createNoopTracer(), nil
	}

	createTraceCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	exporter, err := otlptrace.New(createTraceCtx, client)
	if err != nil {
		return tracing.Tracer{}, err
	}

	serviceVersion := pkg.Version
	if serviceVersion == "" {
		serviceVersion = "dev"
	}
	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(serviceVersion),
	)

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
	)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

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
