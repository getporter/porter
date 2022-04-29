package pluginstore

import (
	"context"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/plugins/proto"
)

var _ plugins.SecretsProtocol = &GClient{}

// GClient is a gRPC implementation of the storage client.
type GClient struct {
	client proto.SecretsProtocolClient
}

func NewClient(client proto.SecretsProtocolClient) *GClient {
	return &GClient{client}
}

func (m *GClient) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	req := &proto.ResolveRequest{
		KeyName:  keyName,
		KeyValue: keyValue,
	}

	resp, err := m.client.Resolve(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func (m *GClient) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	req := &proto.CreateRequest{
		KeyName:  keyName,
		KeyValue: keyValue,
		Value:    value,
	}
	_, err := m.client.Create(ctx, req)
	return err
}

// GServer is a gRPC wrapper around a SecretsProtocol plugin
type GServer struct {
	c    *portercontext.Context
	impl plugins.SecretsProtocol
	proto.UnsafeSecretsProtocolServer
}

func NewServer(c *portercontext.Context, impl plugins.SecretsProtocol) *GServer {
	return &GServer{c: c, impl: impl}
}

func (m *GServer) Resolve(ctx context.Context, request *proto.ResolveRequest) (*proto.ResolveResponse, error) {
	value, err := m.impl.Resolve(ctx, request.KeyName, request.KeyValue)
	if err != nil {
		return nil, err
	}
	return &proto.ResolveResponse{Value: value}, nil
}

func (m *GServer) Create(ctx context.Context, request *proto.CreateRequest) (*proto.CreateResponse, error) {
	err := m.impl.Create(ctx, request.KeyName, request.KeyValue, request.Value)
	if err != nil {
		return nil, err
	}
	return &proto.CreateResponse{}, nil
}
