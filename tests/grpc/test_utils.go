package grpc

import (
	"context"
	"fmt"
	"net"
	"testing"

	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/config"
	pServer "get.porter.sh/porter/pkg/grpc/portergrpc"
	"get.porter.sh/porter/pkg/porter"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	//"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/test/bufconn"
)

type TestPorterGRPCServer struct {
	TestPorter       *porter.TestPorter
	TestPorterConfig *config.TestConfig
}

var lis *bufconn.Listener

func NewTestGRPCServer(t *testing.T) (*TestPorterGRPCServer, error) {
	srv := &TestPorterGRPCServer{
		TestPorter:       porter.NewTestPorter(t),
		TestPorterConfig: config.NewTestConfig(t),
	}
	return srv, nil
}

func (s *TestPorterGRPCServer) newTestInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = pServer.AddPorterConnectionToContext(s.TestPorter.Porter, ctx)
	h, err := handler(ctx, req)
	return h, err
}

func (s *TestPorterGRPCServer) ListenAndServe() *grpc.Server {
	bufSize := 1024 * 1024
	lis = bufconn.Listen(bufSize)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			s.newTestInterceptor,
		)))

	healthServer := health.NewServer()
	reflection.Register(srv)
	grpc_health_v1.RegisterHealthServer(srv, healthServer)
	pSvc, err := pServer.NewPorterServer(s.TestPorter.Config)
	if err != nil {
		panic(err)
	}
	pGRPC.RegisterPorterServer(srv, pSvc)
	healthServer.SetServingStatus("test-health", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(fmt.Errorf("failed to serve: %w", err))
		}
	}()
	return srv
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}
