package portergrpc

import (
	"context"

	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/secrets"
	secretsplugin "get.porter.sh/porter/pkg/secrets/pluginstore"
	"get.porter.sh/porter/pkg/signing"
	signingplugin "get.porter.sh/porter/pkg/signing/pluginstore"
	"get.porter.sh/porter/pkg/storage"
	storageplugin "get.porter.sh/porter/pkg/storage/pluginstore"
	"google.golang.org/grpc"
)

// PorterServer defines the struct for managing a porter GRPC server
type PorterServer struct {
	PorterConfig *config.Config
	pGRPC.UnimplementedPorterServer
}

// NewPorterServer creates a new instance of the PorterServer for a config
func NewPorterServer(cfg *config.Config) (*PorterServer, error) {
	return &PorterServer{PorterConfig: cfg}, nil
}

// NewConnectionInterceptor creates a middleware interceptor for the GRPC server that manages creating a porter connection for each requested RPC stream.
// If the connection is unable to be created for the RPC then the RPC fails, otherwise the connection is added to the RPC context and the next handler in the
// chain is called
func (s *PorterServer) NewConnectionInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	storage := storage.NewPluginAdapter(storageplugin.NewStore(s.PorterConfig))
	secretStorage := secrets.NewPluginAdapter(secretsplugin.NewStore(s.PorterConfig))
	signer := signing.NewPluginAdapter(signingplugin.NewSigner(s.PorterConfig))
	p := porter.NewFor(s.PorterConfig, storage, secretStorage, signer)
	if _, err := p.Connect(ctx); err != nil {
		return nil, err
	}
	defer p.Close()

	ctx = AddPorterConnectionToContext(p, ctx)
	return handler(ctx, req)
}
