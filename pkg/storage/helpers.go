package storage

import (
	"context"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
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

func (s TestStore) Teardown() error {
	return s.testPlugin.Close(context.Background())
}
