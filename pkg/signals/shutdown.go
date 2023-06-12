package signals

import (
	"context"
	"time"

	"get.porter.sh/porter/pkg/tracing"
	"github.com/spf13/viper"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

type Shutdown struct {
	tracerProvider        *oteltrace.TracerProvider
	serverShutdownTimeout time.Duration
}

func NewShutdown(serverShutdownTimeout time.Duration, ctx context.Context) (*Shutdown, error) {
	srv := &Shutdown{
		serverShutdownTimeout: serverShutdownTimeout,
	}

	return srv, nil
}

func (s *Shutdown) Graceful(stopCh <-chan struct{}, grpcServer *grpc.Server, ctx context.Context) {
	ctx, log := tracing.StartSpan(ctx)

	// wait for SIGTERM or SIGINT
	<-stopCh
	ctx, cancel := context.WithTimeout(ctx, s.serverShutdownTimeout)
	defer cancel()

	// wait for Kubernetes readiness probe to remove this instance from the load balancer
	// the readiness check interval must be lower than the timeout
	if viper.GetString("level") != "debug" {
		time.Sleep(3 * time.Second)
	}

	// stop OpenTelemetry tracer provider
	if s.tracerProvider != nil {
		if err := s.tracerProvider.Shutdown(ctx); err != nil {
			log.Warnf("stopping tracer provider: ", err)
		}
	}

	// determine if the GRPC was started
	if grpcServer != nil {
		log.Infof("Shutting down GRPC server: ", s.serverShutdownTimeout)
		grpcServer.GracefulStop()
	}

}
