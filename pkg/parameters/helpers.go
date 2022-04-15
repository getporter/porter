package parameters

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/secrets"

	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage"
	"github.com/carolynvs/aferox"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

var _ Provider = &TestParameterProvider{}

type TestParameterProvider struct {
	*ParameterStore

	T *testing.T
	// TestSecrets allows you to set up secrets for unit testing
	TestSecrets   secrets.TestSecretsProvider
	TestDocuments storage.Store
}

func NewTestParameterProvider(t *testing.T) *TestParameterProvider {
	tc := portercontext.NewTestContext(t)
	testStore := storage.NewTestStore(tc)
	testSecrets := secrets.NewTestSecretsProvider()
	return NewTestParameterProviderFor(t, testStore, testSecrets)
}

func NewTestParameterProviderFor(t *testing.T, testStore storage.Store, testSecrets secrets.TestSecretsProvider) *TestParameterProvider {
	return &TestParameterProvider{
		T:             t,
		TestDocuments: testStore,
		TestSecrets:   testSecrets,
		ParameterStore: &ParameterStore{
			Documents: testStore,
			Secrets:   testSecrets,
		},
	}
}

type hasTeardown interface {
	Teardown() error
}

func (p TestParameterProvider) Teardown() error {
	// sometimes we are testing with a mock that needs to be released at the end of the test
	if ts, ok := p.TestDocuments.(hasTeardown); ok {
		return ts.Teardown()
	} else {
		return p.TestDocuments.Close(context.Background())
	}
}

// Load a ParameterSet from a test file at a given path.
//
// It does not load the individual parameters.
func (p TestParameterProvider) Load(path string) (ParameterSet, error) {
	fs := aferox.NewAferox(".", afero.NewOsFs())
	var pset ParameterSet
	err := encoding.UnmarshalFile(fs, path, &pset)

	return pset, errors.Wrapf(err, "error reading %s as a parameter set", path)
}

func (p TestParameterProvider) AddTestParameters(path string) {
	ps, err := p.Load(path)
	if err != nil {
		p.T.Fatal(errors.Wrapf(err, "could not read test parameters from %s", path))
	}

	err = p.ParameterStore.InsertParameterSet(context.Background(), ps)
	if err != nil {
		p.T.Fatal(errors.Wrap(err, "could not load test parameters"))
	}
}

func (p TestParameterProvider) AddTestParametersDirectory(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		p.T.Fatal(errors.Wrapf(err, "could not list test directory %s", dir))
	}

	for _, fi := range files {
		path := filepath.Join(dir, fi.Name())
		p.AddTestParameters(path)
	}
}
