package storage

import (
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
)

var _ Store = TestStore{}

type TestStore struct {
	PluginAdapter
	testPlugin *testplugin.TestStoragePlugin
}

// NewTestStore creates a store suitable for unit tests.
func NewTestStore(tc *portercontext.TestContext) TestStore {
	testPlugin := testplugin.NewTestStoragePlugin(tc)
	return TestStore{
		PluginAdapter: NewPluginAdapter(testPlugin),
		testPlugin:    testPlugin,
	}
}

func (s TestStore) Teardown() error {
	return s.testPlugin.Teardown()
}
