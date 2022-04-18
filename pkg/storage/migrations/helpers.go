package migrations

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
)

type TestManager struct {
	*Manager

	testStore *storage.TestStore
}

func NewTestManager(c *config.TestConfig) *TestManager {
	testStore := storage.NewTestStore(c)
	return &TestManager{
		Manager: NewManager(c.Config, testStore),
	}
}

func (m *TestManager) Teardown() error {
	return m.testStore.Teardown()
}

// SetSchema allows tests to pre-emptively set the schema document.
func (m *TestManager) SetSchema(schema storage.Schema) {
	m.schema = schema
	m.initialized = true
}
