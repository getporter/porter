package portergrpc

import (
	"context"

	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/secrets"
	secretsplugin "get.porter.sh/porter/pkg/secrets/pluginstore"
	"get.porter.sh/porter/pkg/storage"
	storageplugin "get.porter.sh/porter/pkg/storage/pluginstore"
	"google.golang.org/grpc"
)

type PorterServer struct {
	PorterConfig *config.Config
	pGRPC.UnimplementedPorterServer
}

func NewPorterServer(cfg *config.Config) (*PorterServer, error) {
	return &PorterServer{PorterConfig: cfg}, nil
}

func (s *PorterServer) NewConnectionInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	storage := storage.NewPluginAdapter(storageplugin.NewStore(s.PorterConfig))
	secretStorage := secrets.NewPluginAdapter(secretsplugin.NewStore(s.PorterConfig))
	p := porter.NewFor(s.PorterConfig, storage, secretStorage)
	_, err := p.Connect(ctx)
	if err != nil {
		return nil, err
	}
	defer p.Close()

	ctx = AddPorterConnectionToContext(p, ctx)
	h, err := handler(ctx, req)
	return h, err
}
