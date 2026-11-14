package storage

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
	"github.com/cnabio/cnab-go/secrets/host"
)

type TestStore struct {
	testPlugin *testplugin.TestStoragePlugin
	BackendClient
}

func NewTestStore(tc *config.TestConfig) TestStore {
	testPlugin := testplugin.NewTestStoragePlugin(tc.TestContext)
	return TestStore{
		testPlugin:    testPlugin,
		BackendClient: NewBackendClient(testPlugin),
	}
}

func (s TestStore) Close() error {
	return s.testPlugin.Close()
}

func isHandledByHostPlugin(strategy string) bool {
	return strategy == host.SourceCommand || strategy == host.SourceEnv || strategy == host.SourcePath || strategy == host.SourceValue
}
