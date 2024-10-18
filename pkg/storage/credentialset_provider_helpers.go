package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage/sql/migrate"

	"github.com/carolynvs/aferox"
	"github.com/robinbraemer/devroach"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var _ CredentialSetProvider = &TestCredentialSetProvider{}

type TestCredentialSetProvider struct {
	CredentialSetProvider
	Name string

	T *testing.T
	// TestSecrets allows you to set up secrets for unit testing
	TestSecrets secrets.TestSecretsProvider
}

func NewTestCredentialProvider(t *testing.T) *TestCredentialSetProvider {
	tc := config.NewTestConfig(t)
	testStore := NewTestStore(tc)
	testSecrets := secrets.NewTestSecretsProvider()
	t.Cleanup(func() {
		_ = testStore.Close()
		_ = testSecrets.Close()
	})
	return NewTestCredentialProviderFor(t, testStore, testSecrets)
}

func NewTestCredentialProviderSQL(t *testing.T) *TestCredentialSetProvider {
	db := devroach.NewPoolT(t, nil)
	gormDB, err := gorm.Open(postgres.Open(db.Config().ConnString()), &gorm.Config{})
	require.NoError(t, err)
	err = migrate.MigrateDB(context.Background(), gormDB)
	require.NoError(t, err)

	testSecrets := secrets.NewTestSecretsProvider()
	return &TestCredentialSetProvider{
		T:                     t,
		TestSecrets:           testSecrets,
		CredentialSetProvider: NewCredentialStoreSQL(gormDB, testSecrets),
		Name:                  "PostgreSQL/CockroachDB", // will be updated when supporting other SQL dialects
	}
}

func NewTestCredentialProviderFor(t *testing.T, testStore Store, testSecrets secrets.TestSecretsProvider) *TestCredentialSetProvider {
	return &TestCredentialSetProvider{
		T:           t,
		TestSecrets: testSecrets,
		CredentialSetProvider: &CredentialStore{
			Documents: testStore,
			Secrets:   testSecrets,
		},
		Name: "TestStore/MongoDB",
	}
}

// Load a CredentialSet from a test file at a given path.
//
// It does not load the individual credentials.
func (p TestCredentialSetProvider) Load(path string) (CredentialSet, error) {
	fs := aferox.NewAferox(".", afero.NewOsFs())
	var cset CredentialSet
	err := encoding.UnmarshalFile(fs, path, &cset)
	if err != nil {
		return cset, fmt.Errorf("error reading %s as a credential set: %w", path, err)
	}

	return cset, nil
}

func (p TestCredentialSetProvider) AddTestCredentials(path string) {
	cs, err := p.Load(path)
	if err != nil {
		p.T.Fatal(fmt.Errorf("could not read test credentials from %s: %w", path, err))
	}

	err = p.CredentialSetProvider.InsertCredentialSet(context.Background(), cs)
	if err != nil {
		p.T.Fatal(fmt.Errorf("could not load test credentials into in memory credential storage: %w", err))
	}
}

func (p TestCredentialSetProvider) AddTestCredentialsDirectory(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		p.T.Fatal(fmt.Errorf("could not list test directory %s: %w", dir, err))
	}

	for _, fi := range files {
		path := filepath.Join(dir, fi.Name())
		p.AddTestCredentials(path)
	}
}

func (p TestCredentialSetProvider) AddSecret(key string, value string) {
	err := p.TestSecrets.Create(context.Background(), secrets.SourceSecret, key, value)
	if err != nil {
		p.T.Fatal(fmt.Errorf("could not create secret %s: %w", key, err))
	}
}
