package pluginstore

import (
	"context"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/signing/plugins"
	"get.porter.sh/porter/pkg/signing/plugins/proto"
)

var _ plugins.SigningProtocol = &GClient{}

// GClient is a gRPC implementation of the signing client.
type GClient struct {
	client proto.SigningProtocolClient
}

func NewClient(client proto.SigningProtocolClient) *GClient {
	return &GClient{client}
}

func (m *GClient) Sign(ctx context.Context, ref string) error {
	req := &proto.SignRequest{
		Ref: ref,
	}

	_, err := m.client.Sign(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (m *GClient) Verify(ctx context.Context, ref string) error {
	req := &proto.VerifyRequest{
		Ref: ref,
	}
	_, err := m.client.Verify(ctx, req)
	return err
}

func (m *GClient) Connect(ctx context.Context) error {
	req := &proto.ConnectRequest{}
	_, err := m.client.Connect(ctx, req)
	return err
}

// GServer is a gRPC wrapper around a SecretsProtocol plugin
type GServer struct {
	c    *portercontext.Context
	impl plugins.SigningProtocol
	proto.UnsafeSigningProtocolServer
}

func NewServer(c *portercontext.Context, impl plugins.SigningProtocol) *GServer {
	return &GServer{c: c, impl: impl}
}

func (m *GServer) Sign(ctx context.Context, request *proto.SignRequest) (*proto.SignResponse, error) {
	err := m.impl.Sign(ctx, request.Ref)
	if err != nil {
		return nil, err
	}
	return &proto.SignResponse{}, nil
}

func (m *GServer) Verify(ctx context.Context, request *proto.VerifyRequest) (*proto.VerifyResponse, error) {
	err := m.impl.Verify(ctx, request.Ref)
	if err != nil {
		return nil, err
	}
	return &proto.VerifyResponse{}, nil
}

func (m *GServer) Connect(ctx context.Context, request *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	err := m.impl.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &proto.ConnectResponse{}, nil
}
