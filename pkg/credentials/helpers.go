package credentials

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	inmemorysecrets "get.porter.sh/porter/pkg/secrets/in-memory"
	inmemorystorage "get.porter.sh/porter/pkg/storage/in-memory"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/pkg/errors"
)

var _ CredentialProvider = &TestCredentialProvider{}

type TestCredentialProvider struct {
	T          *testing.T
	TestConfig *config.TestConfig
	*CredentialStorage
}

func NewTestCredentialProvider(t *testing.T, tc *config.TestConfig) TestCredentialProvider {
	backingSecrets := inmemorysecrets.NewStore()
	backingCreds := inmemorystorage.NewStore()
	credStore := credentials.NewCredentialStore(backingCreds)
	return TestCredentialProvider{
		T:          t,
		TestConfig: tc,
		CredentialStorage: &CredentialStorage{
			CredentialsStore: &credStore,
			SecretsStore:     secrets.NewSecretStore(backingSecrets),
		},
	}
}

func (p *TestCredentialProvider) AddTestCredentials(path string) {
	cs, err := credentials.Load(path)
	if err != nil {
		p.T.Fatal(errors.Wrapf(err, "could not read test credentials from %s", path))
	}

	err = p.CredentialStorage.Save(*cs)
	if err != nil {
		p.T.Fatal(errors.Wrap(err, "could not load test credentials into in memory credential storage"))
	}
}

func (p *TestCredentialProvider) AddTestCredentialsDirectory(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		p.T.Fatal(errors.Wrapf(err, "could not list test directory %s", dir))
	}

	for _, fi := range files {
		path := filepath.Join(dir, fi.Name())
		p.AddTestCredentials(path)
	}
}
