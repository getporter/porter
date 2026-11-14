package pluginstore

import (
	"context"

	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var (
	_ plugin.GRPCPlugin      = PluginV4{}
	_ plugins.StorageBackend = &ClientV4{}
)

// PluginV4 is the go-plugin wrapper for a protocol v4 storage backend.
type PluginV4 struct {
	plugin.Plugin
	impl plugins.StorageBackend
}

// NewPluginV4 wraps a backend so that it can be served as a storage plugin.
func NewPluginV4(impl plugins.StorageBackend) PluginV4 {
	return PluginV4{impl: impl}
}

func (p PluginV4) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	protov4.RegisterStorageServer(s, p.impl)
	return nil
}

func (p PluginV4) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return &ClientV4{client: protov4.NewStorageClient(conn)}, nil
}

// ClientV4 implements plugins.StorageBackend by calling a storage plugin over gRPC.
type ClientV4 struct {
	client protov4.StorageClient
}

func (c *ClientV4) Info(ctx context.Context, req *protov4.InfoRequest) (*protov4.InfoResponse, error) {
	return c.client.Info(ctx, req)
}

func (c *ClientV4) Migrate(ctx context.Context, req *protov4.MigrateRequest) (*protov4.MigrateResponse, error) {
	return c.client.Migrate(ctx, req)
}

func (c *ClientV4) Put(ctx context.Context, req *protov4.PutRequest) (*protov4.PutResponse, error) {
	return c.client.Put(ctx, req)
}

func (c *ClientV4) Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error) {
	return c.client.Get(ctx, req)
}

func (c *ClientV4) List(ctx context.Context, req *protov4.ListRequest) (*protov4.ListResponse, error) {
	return c.client.List(ctx, req)
}

func (c *ClientV4) Remove(ctx context.Context, req *protov4.RemoveRequest) (*protov4.RemoveResponse, error) {
	return c.client.Remove(ctx, req)
}

func (c *ClientV4) RemoveInstallation(ctx context.Context, req *protov4.RemoveInstallationRequest) (*protov4.RemoveInstallationResponse, error) {
	return c.client.RemoveInstallation(ctx, req)
}

func (c *ClientV4) ListRuns(ctx context.Context, req *protov4.ListRunsRequest) (*protov4.ListRunsResponse, error) {
	return c.client.ListRuns(ctx, req)
}

func (c *ClientV4) GetLastRun(ctx context.Context, req *protov4.InstallationRef) (*protov4.GetResponse, error) {
	return c.client.GetLastRun(ctx, req)
}

func (c *ClientV4) GetActiveRuns(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	return c.client.GetActiveRuns(ctx, req)
}

func (c *ClientV4) ListResults(ctx context.Context, req *protov4.ListResultsRequest) (*protov4.ListResponse, error) {
	return c.client.ListResults(ctx, req)
}

func (c *ClientV4) InsertOutputs(ctx context.Context, req *protov4.InsertOutputsRequest) (*protov4.InsertOutputsResponse, error) {
	return c.client.InsertOutputs(ctx, req)
}

func (c *ClientV4) ListOutputs(ctx context.Context, req *protov4.ListOutputsRequest) (*protov4.ListResponse, error) {
	return c.client.ListOutputs(ctx, req)
}

func (c *ClientV4) GetOutput(ctx context.Context, req *protov4.GetOutputRequest) (*protov4.GetResponse, error) {
	return c.client.GetOutput(ctx, req)
}

func (c *ClientV4) GetLastOutputs(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	return c.client.GetLastOutputs(ctx, req)
}

func (c *ClientV4) FindInstallationsReferencing(ctx context.Context, req *protov4.FindReferencingRequest) (*protov4.ListResponse, error) {
	return c.client.FindInstallationsReferencing(ctx, req)
}

func (c *ClientV4) FindDependencyCandidates(ctx context.Context, req *protov4.FindDependencyCandidatesRequest) (*protov4.ListResponse, error) {
	return c.client.FindDependencyCandidates(ctx, req)
}
