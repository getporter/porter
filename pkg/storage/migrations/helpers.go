package migrations

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
)

type TestManager struct {
	*Manager

	testPlugin *testplugin.TestStoragePlugin
}

func NewTestManager(c *config.TestConfig) *TestManager {
	testPlugin := testplugin.NewTestStoragePlugin(c.TestContext)
	return &TestManager{
		testPlugin: testPlugin,
		Manager:    NewManager(c.Config, storage.NewPluginAdapter(c.Context, testPlugin)),
	}
}

func (m *TestManager) Teardown() error {
	return m.testPlugin.Teardown()
}

// SetSchema allows tests to pre-emptively set the schema document.
func (m *TestManager) SetSchema(schema storage.Schema) {
	m.schema = schema
	m.initialized = true
}
