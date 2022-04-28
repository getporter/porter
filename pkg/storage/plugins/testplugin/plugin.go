package testplugin

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"
	"get.porter.sh/porter/tests"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	_ plugins.StorageProtocol = &TestStoragePlugin{}
)

// TestStoragePlugin is a test helper that implements a storage plugin backed by a
// mongodb instance that saves data to a temporary directory.
type TestStoragePlugin struct {
	store    *mongodb.Store
	tc       *portercontext.TestContext
	database string
}

func NewTestStoragePlugin(tc *portercontext.TestContext) *TestStoragePlugin {
	return &TestStoragePlugin{tc: tc}
}

// Connect creates a new database for the current test
// The plugin itself is responsible for ensuring it was called.
// Close is called automatically when the plugin is used by Porter.
func (s *TestStoragePlugin) Connect(ctx context.Context) error {
	if s.store != nil {
		return nil
	}

	s.database = tests.GenerateDatabaseName(s.tc.T.Name())

	// Try to connect to a dev instance of mongo, otherwise run a one off mongo instance
	err := s.useDevDatabase(ctx)
	if err != nil {
		// Didn't find a dev mongo instance, so let's run one just for this test
		return s.runTestDatabase(ctx)
	}

	return s.store.RemoveDatabase(ctx) // Start with a fresh test database
}

func (s *TestStoragePlugin) useDevDatabase(ctx context.Context) error {
	cfg := mongodb.PluginConfig{
		URL:     fmt.Sprintf("mongodb://localhost:27017/%s?connect=direct", s.database),
		Timeout: 10,
	}
	devMongo := mongodb.NewStore(s.tc.Context, cfg)
	err := devMongo.Connect(ctx)
	if err != nil {
		return err
	}

	err = devMongo.Ping(ctx)
	if err != nil {
		return err
	}

	s.store = devMongo
	return nil
}

func (s *TestStoragePlugin) runTestDatabase(ctx context.Context) error {
	testMongo, err := mongodb_docker.EnsureMongoIsRunning(ctx, s.tc.Context, "porter-test-mongodb-plugin", "27017", "", s.database, 10)
	if err != nil {
		return err
	}
	s.store = testMongo
	return nil
}

// Close removes the test database and closes the connection.
func (s *TestStoragePlugin) Close() error {
	if s.store != nil {
		var bigErr *multierror.Error
		bigErr = multierror.Append(bigErr, s.store.RemoveDatabase(context.Background()))
		bigErr = multierror.Append(bigErr, s.store.Close())
		s.store = nil
		return bigErr.ErrorOrNil()
	}
	return nil
}

func (s *TestStoragePlugin) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}
	return s.store.EnsureIndex(ctx, opts)
}

func (s *TestStoragePlugin) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.store.Aggregate(ctx, opts)
}

func (s *TestStoragePlugin) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	if err := s.Connect(ctx); err != nil {
		return 0, err
	}
	return s.store.Count(ctx, opts)
}

func (s *TestStoragePlugin) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.store.Find(ctx, opts)
}

func (s *TestStoragePlugin) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}
	return s.store.Insert(ctx, opts)
}

func (s *TestStoragePlugin) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}
	return s.store.Patch(ctx, opts)
}

func (s *TestStoragePlugin) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}
	return s.store.Remove(ctx, opts)
}

func (s *TestStoragePlugin) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}
	return s.store.Update(ctx, opts)
}
