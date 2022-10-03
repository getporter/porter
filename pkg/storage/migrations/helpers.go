package migrations

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
)

type TestManager struct {
	*Manager

	testStore storage.TestStore
}

func NewTestManager(c *config.TestConfig) *TestManager {
	testStore := storage.NewTestStore(c)
	m := &TestManager{
		testStore: testStore,
		Manager:   NewManager(c.Config, testStore),
	}
	ps := storage.NewTestParameterProvider(c.TestContext.T)
	ss := secrets.NewTestSecretsProvider()
	sanitizer := storage.NewSanitizer(ps, ss)
	m.Initialize(sanitizer)
	return m
}

func (m *TestManager) Close() error {
	return m.testStore.Close()
}

// SetSchema allows tests to pre-emptively set the schema document.
func (m *TestManager) SetSchema(schema storage.Schema) {
	m.schema = schema
	m.initialized = true
}
