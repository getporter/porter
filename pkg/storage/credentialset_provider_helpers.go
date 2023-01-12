package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/carolynvs/aferox"
	"github.com/spf13/afero"
)

var _ CredentialSetProvider = &TestCredentialSetProvider{}

type TestCredentialSetProvider struct {
	*CredentialStore

	T           *testing.T
	TestContext *portercontext.TestContext
	// TestSecrets allows you to set up secrets for unit testing
	TestSecrets secrets.TestSecretsProvider
	TestStorage Store
}

func NewTestCredentialProvider(t *testing.T) *TestCredentialSetProvider {
	tc := config.NewTestConfig(t)
	testStore := NewTestStore(tc)
	testSecrets := secrets.NewTestSecretsProvider()
	return NewTestCredentialProviderFor(t, testStore, testSecrets)
}

func NewTestCredentialProviderFor(t *testing.T, testStore Store, testSecrets secrets.TestSecretsProvider) *TestCredentialSetProvider {
	return &TestCredentialSetProvider{
		T:           t,
		TestContext: portercontext.NewTestContext(t),
		TestSecrets: testSecrets,
		TestStorage: testStore,
		CredentialStore: &CredentialStore{
			Documents: testStore,
			Secrets:   testSecrets,
		},
	}
}

func (p TestCredentialSetProvider) Close() error {
	// sometimes we are testing with a mock that needs to be released at the end of the test
	if closer, ok := p.TestStorage.(io.Closer); ok {
		return closer.Close()
	}
	return nil
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

	err = p.CredentialStore.InsertCredentialSet(context.Background(), cs)
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
	p.TestSecrets.Create(context.Background(), secrets.SourceSecret, key, value)
}
