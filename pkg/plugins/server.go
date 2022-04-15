package plugins

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/hashicorp/go-plugin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"google.golang.org/grpc"
)

// Serve a single named plugin.
func Serve(c *portercontext.Context, interfaceName string, pluginImplementation plugin.Plugin, version int) {
	pluginMap := map[int]plugin.PluginSet{
		version: {interfaceName: pluginImplementation},
	}
	ServeMany(c, pluginMap)
}

// ServeMany plugins that the client will select by named interface.
func ServeMany(c *portercontext.Context, pluginMap map[int]plugin.PluginSet) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig:  HandshakeConfig,
		VersionedPlugins: pluginMap,
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			opts = append(opts,
				// These handlers are called from left to right. The right-most handler is the one that calls the actual implementation
				// the grpc_recovery handler should always be last so that it can recover from a panic, and then the other handlers only get
				// a nice error (created from the panic) to deal with
				grpc.ChainUnaryInterceptor(
					otelgrpc.UnaryServerInterceptor(),
					makeLogUnaryHandler(c),
					makePanicHandler()),
			)
			return grpc.NewServer(opts...)
		},
	})
}

// makeLogUnaryHandler creates a span for each RPC method called
func makeLogUnaryHandler(c *portercontext.Context) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		bags := baggage.FromContext(ctx)
		ctx, rootLog := c.StartRootSpan(ctx, info.FullMethod, attribute.String("baggage", bags.String()))
		defer func() {
			rootLog.EndSpan()
		}()

		resp, err := handler(ctx, req)
		return resp, rootLog.Error(err)
	}
}

// makePanicHandler recovers from a panic, logs the error and returns it
func makePanicHandler() grpc.UnaryServerInterceptor {
	recoveryOpts := grpc_recovery.WithRecoveryHandlerContext(func(ctx context.Context, p interface{}) (err error) {
		rootLog := tracing.LoggerFromContext(ctx)
		return rootLog.Error(fmt.Errorf("%s", p))
	})
	return grpc_recovery.UnaryServerInterceptor(recoveryOpts)
}
