package pluginstore

import (
	"context"
	"net"
	"testing"

	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// fakeBackend implements just enough of the backend to test the transport.
type fakeBackend struct {
	plugins.StorageBackend
}

func (fakeBackend) Info(ctx context.Context, req *protov4.InfoRequest) (*protov4.InfoResponse, error) {
	return &protov4.InfoResponse{ProtocolVersion: plugins.PluginProtocolVersionV4, Backend: "fake"}, nil
}

func (fakeBackend) Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error) {
	if req.Name == "missing" {
		return nil, status.Error(codes.NotFound, "installations not found")
	}
	return &protov4.GetResponse{Document: &protov4.Document{
		Kind: req.Kind, Namespace: req.Namespace, Name: req.Name, Json: []byte(`{"name":"` + req.Name + `"}`),
	}}, nil
}

func TestClientV4_RoundTrip(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	require.NoError(t, PluginV4{impl: fakeBackend{}}.GRPCServer(nil, srv))
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	c, err := PluginV4{}.GRPCClient(context.Background(), nil, conn)
	require.NoError(t, err)
	backend := c.(*ClientV4)

	ctx := context.Background()
	info, err := backend.Info(ctx, &protov4.InfoRequest{})
	require.NoError(t, err)
	assert.Equal(t, uint32(plugins.PluginProtocolVersionV4), info.ProtocolVersion)

	got, err := backend.Get(ctx, &protov4.GetRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "ns", Name: "mybuns"})
	require.NoError(t, err)
	assert.Equal(t, "mybuns", got.Document.Name)
	assert.JSONEq(t, `{"name":"mybuns"}`, string(got.Document.Json))

	_, err = backend.Get(ctx, &protov4.GetRequest{Name: "missing"})
	assert.Equal(t, codes.NotFound, status.Code(err), "error codes must survive the transport")
}
