package pluginstore

import (
	"context"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/sbom/plugins"
	"get.porter.sh/porter/pkg/sbom/plugins/proto"
)

var _ plugins.SBOMGeneratorProtocol = &GClient{}

// GClient is a gRPC implementation of the signing client.
type GClient struct {
	client proto.SBOMGeneratorProtocolClient
}

func NewClient(client proto.SBOMGeneratorProtocolClient) *GClient {
	return &GClient{client}
}

func (m *GClient) Generate(ctx context.Context, ref string, sbomPath string, insecureRegistry bool) error {
	req := &proto.GenerateRequest{
		BundleRef:        ref,
		SbomPath:         sbomPath,
		InsecureRegistry: insecureRegistry,
	}

	_, err := m.client.Generate(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (m *GClient) Connect(ctx context.Context) error {
	req := &proto.SBOMConnectRequest{}
	_, err := m.client.Connect(ctx, req)
	return err
}

// GServer is a gRPC wrapper around a SecretsProtocol plugin
type GServer struct {
	c    *portercontext.Context
	impl plugins.SBOMGeneratorProtocol
	proto.UnsafeSBOMGeneratorProtocolServer
}

func NewServer(c *portercontext.Context, impl plugins.SBOMGeneratorProtocol) *GServer {
	return &GServer{c: c, impl: impl}
}

func (m *GServer) Generate(ctx context.Context, request *proto.GenerateRequest) (*proto.GenerateResponse, error) {
	err := m.impl.Generate(ctx, request.BundleRef, request.SbomPath, request.InsecureRegistry)
	if err != nil {
		return nil, err
	}
	return &proto.GenerateResponse{}, nil
}

func (m *GServer) Connect(ctx context.Context, request *proto.SBOMConnectRequest) (*proto.SBOMConnectResponse, error) {
	err := m.impl.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &proto.SBOMConnectResponse{}, nil
}
