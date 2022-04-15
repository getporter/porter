package storage

import (
	"context"
	"fmt"
	"net"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/proto"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: test helpers in grpc.go
func TestNewStructList(t *testing.T) {
	t.Skip()
}

func TestGRPCTelemetry(t *testing.T) {
	c := portercontext.NewTestContext(t)
	store := testplugin.NewTestStoragePlugin(c)

	server := pluginstore.NewServer(c.Context, store)
	addr := fmt.Sprintf("localhost:")
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	proto.RegisterStorageProtocolServer(grpcServer, server)
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := pluginstore.NewClient(proto.NewStorageProtocolClient(conn))

	opts := plugins.EnsureIndexOptions{Indices: []plugins.Index{
		{Collection: "foo", Keys: bson.D{{"name", 1}}}}}
	err = client.EnsureIndex(context.Background(), opts)
	require.NoError(t, err)
}

/*
1. Finish tracing in mongo plugin
2. ensure we are tracing which plugin is running, including version
3. check that all functions trace (e.g. find, ensure index)
4. look for todo comments
5. check for unnecessary changes to our contracts/files
6. add grpc tests to make sure data goes around properly?
7. add otel test with mock tracer?
8. add logger for plugin that emits hclog
9. emit the correlation id for the traces emitted in a plugin before it handles  message
10. how will net/rpc work? do we keep it? (less to test if we don't)
11. update existing plugins to use grpc, telemetry
12. do secrets
make sure that the secrets plugins for azure and hashicorp are updated
fork mchora's helm3
remove net/rpc
look for context.TODO()
make sure code gen for protobuf is happening, and tools automatically installed
*/
