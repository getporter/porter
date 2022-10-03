package pluginstore

import (
	"context"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/plugins/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var _ plugin.GRPCPlugin = Plugin{}

// Plugin is the shared implementation of a storage plugin wrapper.
type Plugin struct {
	plugin.Plugin
	impl    plugins.SecretsProtocol
	context *portercontext.Context
}

// NewPlugin creates an instance of a storage plugin.
func NewPlugin(c *portercontext.Context, impl plugins.SecretsProtocol) Plugin {
	return Plugin{
		context: c,
		impl:    impl,
	}
}

func (p Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	impl := NewServer(p.context, p.impl)
	proto.RegisterSecretsProtocolServer(s, impl)
	return nil
}

func (p Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return NewClient(proto.NewSecretsProtocolClient(conn)), nil
}
