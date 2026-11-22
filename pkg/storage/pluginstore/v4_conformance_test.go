package pluginstore_test

import (
	"net"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/plugintest"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Storage plugins run in their own process and talk to Porter over gRPC. The
// conformance suite must hold across that boundary too, for example error codes
// such as NotFound and AlreadyExists must survive the trip.
func TestClientV4_Conformance(t *testing.T) {
	plugintest.RunConformance(t, func(t *testing.T) plugins.StorageBackend {
		tc := portercontext.NewTestContext(t)
		backend := testplugin.NewTestStoragePlugin(tc)
		t.Cleanup(func() { backend.Close() })

		lis, err := net.Listen("tcp", "localhost:")
		require.NoError(t, err)
		srv := grpc.NewServer()
		require.NoError(t, pluginstore.NewPluginV4(backend).GRPCServer(nil, srv))
		go func() { _ = srv.Serve(lis) }()
		t.Cleanup(srv.Stop)

		conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		t.Cleanup(func() { conn.Close() })

		client, err := pluginstore.PluginV4{}.GRPCClient(t.Context(), nil, conn)
		require.NoError(t, err)
		return client.(*pluginstore.ClientV4)
	})
}
