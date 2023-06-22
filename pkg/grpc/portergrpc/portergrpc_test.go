package portergrpc

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewConnectionInterceptorCallsNextHandlerInTheChainWithThePorterConnectionInTheContext(t *testing.T) {
	cfg := config.NewTestConfig(t)
	srv := PorterServer{PorterConfig: cfg.Config}
	parentUnaryInfo := &grpc.UnaryServerInfo{FullMethod: "SomeService.StreamMethod"}
	input := "input"
	testHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		p, err := GetPorterConnectionFromContext(ctx)
		return p, err
	}
	ctx := context.Background()
	chain := grpc_middleware.ChainUnaryServer(srv.NewConnectionInterceptor)
	newP, err := chain(ctx, input, parentUnaryInfo, testHandler)
	assert.Nil(t, err)
	assert.NotNil(t, newP)
	assert.IsType(t, &porter.Porter{}, newP)
}
