package testplugin

import (
	"fmt"
	"runtime"

	"get.porter.sh/porter/pkg/context"
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
	*mongodb.Store

	tc       *context.TestContext
	database string
}

func NewTestStoragePlugin(tc *context.TestContext) *TestStoragePlugin {
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
func (s *TestStoragePlugin) Setup() error {
	if s.Store != nil {
		return nil
	}

	s.database = tests.GenerateDatabaseName(s.tc.T.Name())

	// Try to connect to a dev instance of mongo, otherwise run a one off mongo instance
	err := s.useDevDatabase()
	if err != nil {
		// Didn't find a dev mongo instance, so let's run one just for this test
		return s.runTestDatabase()
	}

	return s.Store.RemoveDatabase() // Start with a fresh test database
}

func (s *TestStoragePlugin) useDevDatabase() error {
	cfg := mongodb.PluginConfig{
		URL: fmt.Sprintf("mongodb://localhost:27017/%s?connect=direct", s.database),
	}
	devMongo := mongodb.NewStore(s.tc.Context, cfg)
	err := devMongo.Connect()
	if err != nil {
		return err
	}

	err = devMongo.Ping()
	if err != nil {
		return err
	}

	s.Store = devMongo
	return nil
}

func (s *TestStoragePlugin) runTestDatabase() error {
	testMongo, err := mongodb_docker.EnsureMongoIsRunning(s.tc.Context, "porter-test-mongodb-plugin", "27017", "", s.database, 10)
	if err != nil {
		return err
	}
	s.Store = testMongo
	return nil
}

// Teardown stops the test mongo instance and cleans up any temporary files.
func (s *TestStoragePlugin) Teardown() error {
	if s.Store != nil {
		s.Store.RemoveDatabase()
		return s.Close()
	}
	return nil
}

// Connect sets up the test mongo instance if necessary.
func (s *TestStoragePlugin) Connect() error {
	return s.Setup()
}

// Close the connection to the database.
func (s *TestStoragePlugin) Close() error {
	err := s.Store.Close()
	s.Store = nil
	return err
}
