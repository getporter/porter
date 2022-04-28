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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRoundTripDataOverGRPC(t *testing.T) {
	// Just check that we can round trip data through our storage grpc service
	c := portercontext.NewTestContext(t)
	store := testplugin.NewTestStoragePlugin(c)
	ctx := context.Background()

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

	// Add an index to support filtering
	const collection = "things"
	err = client.EnsureIndex(ctx, plugins.EnsureIndexOptions{Indices: []plugins.Index{
		{Collection: collection, Keys: bson.D{{"namespace", 1}, {"name", 1}}},
	}})
	require.NoError(t, err)

	thing1 := bson.M{"name": "Thing1", "namespace": "dev"}
	err = client.Insert(ctx, plugins.InsertOptions{
		Collection: collection,
		Documents: []bson.M{
			thing1,
			{"name": "Thing2", "namespace": "staging"},
		},
	})
	require.NoError(t, err)

	results, err := client.Find(ctx, plugins.FindOptions{
		Collection: collection,
		Filter:     bson.M{"namespace": "dev"},
		Select:     bson.D{{"name", 1}, {"namespace", 1}, {"_id", 0}},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	var gotThing1 bson.M
	require.NoError(t, bson.Unmarshal(results[0], &gotThing1))
	assert.Equal(t, thing1, gotThing1)
}
