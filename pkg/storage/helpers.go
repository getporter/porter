package storage

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
	"github.com/cnabio/cnab-go/secrets/host"
)

var _ Store = TestStore{}

type TestStore struct {
	testPlugin *testplugin.TestStoragePlugin
	PluginAdapter
}

func NewTestStore(tc *config.TestConfig) TestStore {
	testPlugin := testplugin.NewTestStoragePlugin(tc.TestContext)
	return TestStore{
		testPlugin:    testPlugin,
		PluginAdapter: NewPluginAdapter(testPlugin),
	}
}

func (s TestStore) Close() error {
	return s.testPlugin.Close()
}

func isHandledByHostPlugin(strategy string) bool {
	return strategy == host.SourceCommand || strategy == host.SourceEnv || strategy == host.SourcePath || strategy == host.SourceValue
}
