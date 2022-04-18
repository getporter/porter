package storage

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
	"get.porter.sh/porter/pkg/storage/pluginstore"
)

var _ Store = TestStore{}

type TestStore struct {
	testPlugin *testplugin.TestStoragePlugin
	PluginAdapter
}

func NewTestStore(tc *config.TestConfig) TestStore {
	testPlugin := testplugin.NewTestStoragePlugin(tc.TestContext)
	store := pluginstore.NewStore(tc.Config)
	store.SetPlugin(testPlugin)
	return TestStore{
		testPlugin:    testPlugin,
		PluginAdapter: NewPluginAdapter(store),
	}
}

func (s TestStore) Teardown() error {
	return s.testPlugin.Teardown()
}
