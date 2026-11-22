package testplugin

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"get.porter.sh/porter/tests"
	"github.com/hashicorp/go-multierror"
)

var (
	_ plugins.StorageBackend = &TestStoragePlugin{}
)

// TestStoragePlugin is a test helper that implements a storage plugin backed by a
// mongodb instance that saves data to a temporary directory.
type TestStoragePlugin struct {
	store    *mongodb.Store
	backend  *mongodb.Backend
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
	defer func() {
		if s.store != nil {
			s.backend = mongodb.NewBackend(s.store)
		}
	}()

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
		s.backend = nil
		return bigErr.ErrorOrNil()
	}
	return nil
}

func (s *TestStoragePlugin) Info(ctx context.Context, req *protov4.InfoRequest) (*protov4.InfoResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Info(ctx, req)
}

func (s *TestStoragePlugin) Migrate(ctx context.Context, req *protov4.MigrateRequest) (*protov4.MigrateResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Migrate(ctx, req)
}

func (s *TestStoragePlugin) Put(ctx context.Context, req *protov4.PutRequest) (*protov4.PutResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Put(ctx, req)
}

func (s *TestStoragePlugin) Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Get(ctx, req)
}

func (s *TestStoragePlugin) List(ctx context.Context, req *protov4.ListRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.List(ctx, req)
}

func (s *TestStoragePlugin) Remove(ctx context.Context, req *protov4.RemoveRequest) (*protov4.RemoveResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Remove(ctx, req)
}

func (s *TestStoragePlugin) RemoveInstallation(ctx context.Context, req *protov4.RemoveInstallationRequest) (*protov4.RemoveInstallationResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.RemoveInstallation(ctx, req)
}

func (s *TestStoragePlugin) ListRuns(ctx context.Context, req *protov4.ListRunsRequest) (*protov4.ListRunsResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.ListRuns(ctx, req)
}

func (s *TestStoragePlugin) GetLastRun(ctx context.Context, req *protov4.InstallationRef) (*protov4.GetResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.GetLastRun(ctx, req)
}

func (s *TestStoragePlugin) GetActiveRuns(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.GetActiveRuns(ctx, req)
}

func (s *TestStoragePlugin) ListResults(ctx context.Context, req *protov4.ListResultsRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.ListResults(ctx, req)
}

func (s *TestStoragePlugin) InsertOutputs(ctx context.Context, req *protov4.InsertOutputsRequest) (*protov4.InsertOutputsResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.InsertOutputs(ctx, req)
}

func (s *TestStoragePlugin) ListOutputs(ctx context.Context, req *protov4.ListOutputsRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.ListOutputs(ctx, req)
}

func (s *TestStoragePlugin) GetOutput(ctx context.Context, req *protov4.GetOutputRequest) (*protov4.GetResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.GetOutput(ctx, req)
}

func (s *TestStoragePlugin) GetLastOutputs(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.GetLastOutputs(ctx, req)
}

func (s *TestStoragePlugin) FindInstallationsReferencing(ctx context.Context, req *protov4.FindReferencingRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.FindInstallationsReferencing(ctx, req)
}

func (s *TestStoragePlugin) FindDependencyCandidates(ctx context.Context, req *protov4.FindDependencyCandidatesRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.FindDependencyCandidates(ctx, req)
}
