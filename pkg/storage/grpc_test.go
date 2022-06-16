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
		{Collection: collection, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "name", Value: 1}}},
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
		Select:     bson.D{{Key: "name", Value: 1}, {Key: "namespace", Value: 1}, {Key: "_id", Value: 0}},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	var gotThing1 bson.M
	require.NoError(t, bson.Unmarshal(results[0], &gotThing1))
	assert.Equal(t, thing1, gotThing1)

	opts := plugins.EnsureIndexOptions{
		Indices: []plugins.Index{
			// query most recent outputs by run (porter installation run show, when we list outputs)
			{Collection: CollectionOutputs, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "installation", Value: 1}, {Key: "-resultId", Value: 1}}},
			// query outputs by result (list)
			{Collection: CollectionOutputs, Keys: bson.D{{Key: "resultId", Value: 1}, {Key: "name", Value: 1}}, Unique: true},
			// query most recent outputs by name for an installation
			{Collection: CollectionOutputs, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "installation", Value: 1}, {Key: "name", Value: 1}, {Key: "-resultId", Value: 1}}},
		},
	}

	err = client.EnsureIndex(ctx, opts)
	require.NoError(t, err)

	err = client.Insert(ctx, plugins.InsertOptions{
		Collection: CollectionOutputs,
		Documents: []bson.M{{"namespace": "dev", "installation": "test", "name": "thing1", "resultId": "111"},
			{"namespace": "dev", "installation": "test", "name": "thing2", "resultId": "111"},
			{"namespace": "dev", "installation": "test", "name": "thing2", "resultId": "222"},
		},
	})
	require.NoError(t, err)

	aggregateResults, err := client.Aggregate(ctx, plugins.AggregateOptions{
		Collection: CollectionOutputs,
		Pipeline: []bson.D{
			// List outputs by installation
			{{Key: "$match", Value: bson.M{
				"namespace":    "dev",
				"installation": "test",
			}}},
			// Reverse sort them (newest on top)
			{{Key: "$sort", Value: bson.D{
				{Key: "namespace", Value: 1},
				{Key: "installation", Value: 1},
				{Key: "name", Value: 1},
				{Key: "resultId", Value: -1},
			}}},
			// Group them by output name and select the last value for each output
			{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$name"},
				{Key: "lastOutput", Value: bson.M{"$first": "$$ROOT"}},
			}}},
		},
	})
	require.NoError(t, err)
	require.Len(t, aggregateResults, 2)
	// make sure the group function is picking the most recent output value with the same name
	require.Contains(t, aggregateResults[0].Lookup("lastOutput").String(), "222")
}
