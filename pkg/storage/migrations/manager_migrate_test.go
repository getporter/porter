package migrations

import (
	"context"
	"errors"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// emptyBackend behaves like a brand-new database: it has no tables until it is
// migrated, so any other call before Migrate fails.
type emptyBackend struct {
	plugins.StorageBackend

	migrated    bool
	migrateErr  error
	migrateCall int
	calls       []string
}

func (b *emptyBackend) Migrate(ctx context.Context, req *protov4.MigrateRequest) (*protov4.MigrateResponse, error) {
	b.migrateCall++
	b.calls = append(b.calls, "Migrate")
	if b.migrateErr != nil {
		return nil, b.migrateErr
	}
	b.migrated = true
	return &protov4.MigrateResponse{}, nil
}

func (b *emptyBackend) Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error) {
	b.calls = append(b.calls, "Get")
	if !b.migrated {
		return nil, status.Error(codes.Internal, "relation does not exist")
	}
	return nil, status.Error(codes.NotFound, "schema not found")
}

func (b *emptyBackend) List(ctx context.Context, req *protov4.ListRequest) (*protov4.ListResponse, error) {
	b.calls = append(b.calls, "List")
	if !b.migrated {
		return nil, status.Error(codes.Internal, "relation does not exist")
	}
	return &protov4.ListResponse{}, nil
}

func (b *emptyBackend) Put(ctx context.Context, req *protov4.PutRequest) (*protov4.PutResponse, error) {
	b.calls = append(b.calls, "Put")
	return &protov4.PutResponse{}, nil
}

func newManagerWithoutCache(t *testing.T, b plugins.StorageBackend) *Manager {
	tc := config.NewTestConfig(t)
	// Don't let a cached check from another test skip the connection check
	m := NewManager(tc.Config, b)
	m.allowOutOfDateSchema = true
	return m
}

func TestManager_Connect_MigratesBackendBeforeReadingSchema(t *testing.T) {
	ctx := context.Background()
	b := &emptyBackend{}
	m := newManagerWithoutCache(t, b)

	require.NoError(t, m.Connect(ctx))

	require.NotEmpty(t, b.calls)
	assert.Equal(t, "Migrate", b.calls[0], "the backend must be prepared before porter reads from it")
	assert.Contains(t, b.calls, "Put", "a new database should have its schema written")
}

func TestManager_Connect_RetriesAFailedBackendMigration(t *testing.T) {
	ctx := context.Background()
	b := &emptyBackend{migrateErr: errors.New("database is unreachable")}
	m := newManagerWithoutCache(t, b)

	err := m.Connect(ctx)
	require.ErrorContains(t, err, "could not prepare the storage backend")
	assert.False(t, m.initialized, "a failed connection must not be remembered")

	b.migrateErr = nil
	require.NoError(t, m.Connect(ctx), "the next call should try again")
	assert.Equal(t, 2, b.migrateCall)
	assert.True(t, m.initialized)
}
