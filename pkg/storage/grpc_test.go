package storage

import (
	"context"
	"net"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRoundTripDataOverGRPC(t *testing.T) {
	// Check that we can round trip typed documents through the storage grpc service
	c := portercontext.NewTestContext(t)
	backend := testplugin.NewTestStoragePlugin(c)
	ctx := context.Background()
	defer backend.Close()

	lis, err := net.Listen("tcp", "localhost:")
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	require.NoError(t, pluginstore.NewPluginV4(backend).GRPCServer(nil, grpcServer))
	errs := make(chan error, 1)
	go func() {
		errs <- grpcServer.Serve(lis)
	}()
	defer func() {
		grpcServer.Stop()
		if err := <-errs; err != nil {
			require.NoError(t, err)
		}
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	remote, err := pluginstore.PluginV4{}.GRPCClient(ctx, nil, conn)
	require.NoError(t, err)
	client := NewBackendClient(remote.(*pluginstore.ClientV4))

	// Create the indexes that the queries rely on
	_, err = client.Backend().Migrate(ctx, &protov4.MigrateRequest{})
	require.NoError(t, err)

	inst1 := NewInstallation("dev", "thing1")
	inst2 := NewInstallation("dev", "thing2")
	inst3 := NewInstallation("prod", "thing3")
	require.NoError(t, client.put(ctx, protov4.PutRequest_INSERT, inst1, inst2, inst3))

	t.Run("get", func(t *testing.T) {
		got, err := getDocument[Installation](ctx, client, &protov4.GetRequest{
			Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev", Name: "thing1",
		})
		require.NoError(t, err)
		assert.Equal(t, inst1.ID, got.ID)
	})

	t.Run("get missing", func(t *testing.T) {
		_, err := getDocument[Installation](ctx, client, &protov4.GetRequest{
			Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev", Name: "missing",
		})
		assert.ErrorIs(t, err, ErrNotFound{})
	})

	t.Run("list by namespace", func(t *testing.T) {
		got, err := listDocuments[Installation](ctx, client, ListOptions{Namespace: "dev"}.toListRequest(protov4.Kind_KIND_INSTALLATION))
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "thing1", got[0].Name)
		assert.Equal(t, "thing2", got[1].Name)
	})

	t.Run("list all namespaces", func(t *testing.T) {
		got, err := listDocuments[Installation](ctx, client, ListOptions{Namespace: "*"}.toListRequest(protov4.Kind_KIND_INSTALLATION))
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("duplicate insert", func(t *testing.T) {
		err := client.put(ctx, protov4.PutRequest_INSERT, inst1)
		require.Error(t, err, "the unique index on namespace and name should reject a duplicate")
	})
}
