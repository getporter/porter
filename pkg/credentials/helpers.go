package credentials

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"

	"get.porter.sh/porter/pkg/secrets"

	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage"
	"github.com/carolynvs/aferox"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

var _ Provider = &TestCredentialProvider{}

type TestCredentialProvider struct {
	*CredentialStore

	T           *testing.T
	TestContext *portercontext.TestContext
	// TestSecrets allows you to set up secrets for unit testing
	TestSecrets secrets.TestSecretsProvider
	TestStorage storage.Store
}

func NewTestCredentialProvider(t *testing.T) *TestCredentialProvider {
	tc := config.NewTestConfig(t)
	testStore := storage.NewTestStore(tc)
	testSecrets := secrets.NewTestSecretsProvider()
	return NewTestCredentialProviderFor(t, testStore, testSecrets)
}

func NewTestCredentialProviderFor(t *testing.T, testStore storage.Store, testSecrets secrets.TestSecretsProvider) *TestCredentialProvider {
	return &TestCredentialProvider{
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

type hasTeardown interface {
	Teardown() error
}

func (p TestCredentialProvider) Teardown() error {
	// sometimes we are testing with a mock that needs to be released at the end of the test
	if ts, ok := p.TestStorage.(hasTeardown); ok {
		return ts.Teardown()
	} else {
		return p.TestStorage.Close(context.Background())
	}
}

// Load a CredentialSet from a test file at a given path.
//
// It does not load the individual credentials.
func (p TestCredentialProvider) Load(path string) (CredentialSet, error) {
	fs := aferox.NewAferox(".", afero.NewOsFs())
	var cset CredentialSet
	err := encoding.UnmarshalFile(fs, path, &cset)

	return cset, errors.Wrapf(err, "error reading %s as a credential set", path)
}

func (p TestCredentialProvider) AddTestCredentials(path string) {
	cs, err := p.Load(path)
	if err != nil {
		p.T.Fatal(errors.Wrapf(err, "could not read test credentials from %s", path))
	}

	err = p.CredentialStore.InsertCredentialSet(context.Background(), cs)
	if err != nil {
		p.T.Fatal(errors.Wrap(err, "could not load test credentials into in memory credential storage"))
	}
}

func (p TestCredentialProvider) AddTestCredentialsDirectory(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		p.T.Fatal(errors.Wrapf(err, "could not list test directory %s", dir))
	}

	for _, fi := range files {
		path := filepath.Join(dir, fi.Name())
		p.AddTestCredentials(path)
	}
}
