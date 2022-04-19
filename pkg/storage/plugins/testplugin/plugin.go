package testplugin

import (
	"context"
	"fmt"
	"runtime"

	"go.mongodb.org/mongo-driver/bson"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"
	"get.porter.sh/porter/tests"
)

var (
	_ plugins.StoragePlugin = &TestStoragePlugin{}
)

// TestStoragePlugin is a test helper that implements a storage plugin backed by a
// mongodb instance that saves data to a temporary directory.
type TestStoragePlugin struct {
	store *mongodb.Store

	tc       *portercontext.TestContext
	database string
}

func NewTestStoragePlugin(tc *portercontext.TestContext) *TestStoragePlugin {
	p := &TestStoragePlugin{tc: tc}

	// This is extra insurance that when we are running tests in the debugger
	// that we don't accidentally end the test before calling Close() and then
	// leaking mongodb processes.
	runtime.SetFinalizer(p, func(p *TestStoragePlugin) {
		p.Teardown()
	})

	return p
}

// Setup runs mongodb on an alternate Port
func (s *TestStoragePlugin) Setup(ctx context.Context) error {
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

// Teardown stops the test mongo instance and cleans up any temporary files.
func (s *TestStoragePlugin) Teardown() error {
	ctx := context.TODO()
	if s.store != nil {
		s.store.RemoveDatabase(ctx)
		return s.Close(ctx)
	}
	return nil
}

// Connect sets up the test mongo instance if necessary.
func (s *TestStoragePlugin) Connect(ctx context.Context) error {
	return s.Setup(ctx)
}

// Close the connection to the database.
func (s *TestStoragePlugin) Close(ctx context.Context) error {
	if s.store != nil {
		err := s.store.Close(ctx)
		s.store = nil
		return err
	}
	return nil
}

func (s *TestStoragePlugin) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	return s.store.EnsureIndex(ctx, opts)
}

func (s *TestStoragePlugin) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	return s.store.Aggregate(ctx, opts)
}

func (s *TestStoragePlugin) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	return s.store.Count(ctx, opts)
}

func (s *TestStoragePlugin) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	return s.store.Find(ctx, opts)
}

func (s *TestStoragePlugin) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	return s.store.Insert(ctx, opts)
}

func (s *TestStoragePlugin) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	return s.store.Patch(ctx, opts)
}

func (s *TestStoragePlugin) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	return s.store.Remove(ctx, opts)
}

func (s *TestStoragePlugin) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	return s.store.Update(ctx, opts)
}
