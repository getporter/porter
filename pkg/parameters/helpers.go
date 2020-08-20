package parameters

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	inmemorysecrets "get.porter.sh/porter/pkg/secrets/in-memory"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

var _ ParameterProvider = &TestParameterProvider{}

type TestParameterProvider struct {
	T          *testing.T
	TestConfig *config.TestConfig
	// TestSecrets allows you to set up secrets for unit testing
	TestSecrets *inmemorysecrets.Store
	*ParameterStorage
}

func NewTestParameterProvider(t *testing.T, tc *config.TestConfig) TestParameterProvider {
	backingSecrets := inmemorysecrets.NewStore()
	backingParams := crud.NewBackingStore(crud.NewMockStore())
	paramStore := NewParameterStore(backingParams)
	return TestParameterProvider{
		T:           t,
		TestConfig:  tc,
		TestSecrets: backingSecrets,
		ParameterStorage: &ParameterStorage{
			ParametersStore: paramStore,
			SecretsStore:    secrets.NewSecretStore(backingSecrets),
		},
	}
}

func (p *TestParameterProvider) AddTestParameters(path string) {
	cs, err := Load(path)
	if err != nil {
		p.T.Fatal(errors.Wrapf(err, "could not read test parameters from %s", path))
	}

	err = p.ParameterStorage.Save(cs)
	if err != nil {
		p.T.Fatal(errors.Wrap(err, "could not load test parameters into in memory parameter storage"))
	}
}

func (p *TestParameterProvider) AddTestParametersDirectory(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		p.T.Fatal(errors.Wrapf(err, "could not list test directory %s", dir))
	}

	for _, fi := range files {
		path := filepath.Join(dir, fi.Name())
		p.AddTestParameters(path)
	}
}
