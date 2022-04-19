package pluginstore

import (
	"context"
	"net/rpc"

	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var _ plugin.Plugin = &Plugin{}

// Plugin is the net/rpc implementation of the storage plugin.
type Plugin struct {
	Impl plugins.StorageProtocol
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &Server{Impl: p.Impl}, nil
}

func (Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &Client{client: c}, nil
}

var _ plugin.GRPCPlugin = &GPlugin{}

// GPlugin is the gRPC implementation of the storage plugin.
type GPlugin struct {
	plugin.Plugin
	Impl plugins.StorageProtocol
}

func (p *GPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterStorageProtocolServer(s, &GServer{Impl: p.Impl})
	return nil
}

func (p *GPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GClient{client: proto.NewStorageProtocolClient(c)}, nil
}
