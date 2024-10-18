package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/carolynvs/aferox"
	"github.com/robinbraemer/devroach"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage/sql/migrate"
)

var _ ParameterSetProvider = &TestParameterSetProvider{}

type TestParameterSetProvider struct {
	ParameterSetProvider
	Name string

	T *testing.T
	// TestSecrets allows you to set up secrets for unit testing
	TestSecrets secrets.TestSecretsProvider
}

func NewTestParameterProvider(t *testing.T) *TestParameterSetProvider {
	tc := config.NewTestConfig(t)
	testStore := NewTestStore(tc)
	testSecrets := secrets.NewTestSecretsProvider()
	t.Cleanup(func() {
		_ = testStore.Close()
		_ = testSecrets.Close()
	})
	return NewTestParameterProviderFor(t, testStore, testSecrets)
}

func NewTestParameterProviderSQL(t *testing.T) *TestParameterSetProvider {
	db := devroach.NewPoolT(t, nil)
	gormDB, err := gorm.Open(postgres.Open(db.Config().ConnString()), &gorm.Config{})
	require.NoError(t, err)
	err = migrate.MigrateDB(context.Background(), gormDB)
	require.NoError(t, err)
	testSecrets := secrets.NewTestSecretsProvider()

	return &TestParameterSetProvider{
		T:                    t,
		TestSecrets:          testSecrets,
		ParameterSetProvider: NewParameterStoreSQL(gormDB, testSecrets),
		Name:                 "PostgreSQL/CockroachDB", // will be updated when supporting other SQL dialects
	}
}

func NewTestParameterProviderFor(t *testing.T, testStore Store, testSecrets secrets.TestSecretsProvider) *TestParameterSetProvider {
	return &TestParameterSetProvider{
		T:           t,
		TestSecrets: testSecrets,
		ParameterSetProvider: &ParameterStore{
			Documents: testStore,
			Secrets:   testSecrets,
		},
		Name: "TestStore/MongoDB",
	}
}

// Load a ParameterSet from a test file at a given path.
//
// It does not load the individual parameters.
func (p TestParameterSetProvider) Load(path string) (ParameterSet, error) {
	fs := aferox.NewAferox(".", afero.NewOsFs())
	var pset ParameterSet
	err := encoding.UnmarshalFile(fs, path, &pset)
	if err != nil {
		return pset, fmt.Errorf("error reading %s as a parameter set: %w", path, err)
	}

	return pset, nil
}

func (p TestParameterSetProvider) AddTestParameters(path string) {
	ps, err := p.Load(path)
	if err != nil {
		p.T.Fatal(fmt.Errorf("could not read test parameters from %s: %w", path, err))
	}

	err = p.ParameterSetProvider.InsertParameterSet(context.Background(), ps)
	if err != nil {
		p.T.Fatal(fmt.Errorf("could not load test parameters: %w", err))
	}
}

func (p TestParameterSetProvider) AddTestParametersDirectory(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		p.T.Fatal(fmt.Errorf("could not list test directory %s: %w", dir, err))
	}

	for _, fi := range files {
		path := filepath.Join(dir, fi.Name())
		p.AddTestParameters(path)
	}
}

func (p TestParameterSetProvider) AddSecret(key string, value string) {
	err := p.TestSecrets.Create(context.Background(), secrets.SourceSecret, key, value)
	if err != nil {
		p.T.Fatal(fmt.Errorf("could not create secret %s: %w", key, err))
	}
}
