package migrations

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
)

type TestManager struct {
	*Manager

	testStore storage.TestStore
}

func NewTestManager(c *config.TestConfig) *TestManager {
	testStore := storage.NewTestStore(c)
	return &TestManager{
		testStore: testStore,
		Manager:   NewManager(c.Config, testStore),
	}
}

func (m *TestManager) Close() error {
	return m.testStore.Close()
}

// SetSchema allows tests to pre-emptively set the schema document.
func (m *TestManager) SetSchema(schema storage.Schema) {
	m.schema = schema
	m.initialized = true
}
