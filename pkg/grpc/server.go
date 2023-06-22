package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	pGRPCv1alpha1 "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/config"
	pserver "get.porter.sh/porter/pkg/grpc/portergrpc"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/tracing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	reg = prometheus.NewRegistry()
	// Create some standard server metrics.
	grpcMetrics = grpc_prometheus.NewServerMetrics()

	// Create a customized counter metric.
	customizedCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "demo_server_say_hello_method_handle_count",
		Help: "Total number of RPCs handled on the server.",
	}, []string{"name"})
)

func init() {
	reg.MustRegister(grpcMetrics, customizedCounterMetric)
	customizedCounterMetric.WithLabelValues("Test")
}

type PorterGRPCService struct {
	PorterConfig *config.Config
	opts         *porter.ServiceOptions
	ctx          context.Context
}

func NewServer(ctx context.Context, opts *porter.ServiceOptions) (*PorterGRPCService, error) {
	pCfg := config.New()
	srv := &PorterGRPCService{
		PorterConfig: pCfg,
		opts:         opts,
		ctx:          ctx,
	}
	return srv, nil
}

func (s *PorterGRPCService) ListenAndServe() (*grpc.Server, error) {
	_, log := tracing.StartSpan(s.ctx)
	defer log.EndSpan()
	log.Infof("Starting gRPC on %v", s.opts.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.opts.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %d: %s", s.opts.Port, err)
	}
	httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: fmt.Sprintf("0.0.0.0:%d", 9092)}

	healthServer := health.NewServer()
	psrv, err := pserver.NewPorterServer(s.PorterConfig)
	if err != nil {
		return nil, err
	}
	srv := grpc.NewServer(
		grpc.StreamInterceptor(grpcMetrics.StreamServerInterceptor()),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpcMetrics.UnaryServerInterceptor(),
			psrv.NewConnectionInterceptor),
		),
	)
	reflection.Register(srv)

	pGRPCv1alpha1.RegisterPorterServer(srv, psrv)
	healthServer.SetServingStatus(s.opts.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_prometheus.Register(srv)

	srvErrCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(listener); err != nil {
			srvErrCh <- err
		}
	}()

	select {
	case err := <-srvErrCh:
		if err != nil {
			log.Errorf("failed to serve GRPC listener: %w", err)
			os.Exit(1)
		}
	default:
		// Continue if no error is received yet
	}

	http.Handle("/metrics", promhttp.Handler())

	// Start your http server for prometheus.
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			srvErrCh <- err
		}
	}()
	select {
	case err := <-srvErrCh:
		if err != nil {
			log.Errorf("Unable to start a http server. %w", err)
			os.Exit(1)
		}
	default:
		// Continue if no error is received yet
	}

	grpcMetrics.InitializeMetrics(srv)
	return srv, nil
}
