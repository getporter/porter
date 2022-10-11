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
	"github.com/carolynvs/aferox"
	"github.com/spf13/afero"
)

var _ ParameterSetProvider = &TestParameterSetProvider{}

type TestParameterSetProvider struct {
	*ParameterStore

	T *testing.T
	// TestSecrets allows you to set up secrets for unit testing
	TestSecrets   secrets.TestSecretsProvider
	TestDocuments Store
}

func NewTestParameterProvider(t *testing.T) *TestParameterSetProvider {
	tc := config.NewTestConfig(t)
	testStore := NewTestStore(tc)
	testSecrets := secrets.NewTestSecretsProvider()
	return NewTestParameterProviderFor(t, testStore, testSecrets)
}

func NewTestParameterProviderFor(t *testing.T, testStore Store, testSecrets secrets.TestSecretsProvider) *TestParameterSetProvider {
	return &TestParameterSetProvider{
		T:             t,
		TestDocuments: testStore,
		TestSecrets:   testSecrets,
		ParameterStore: &ParameterStore{
			Documents: testStore,
			Secrets:   testSecrets,
		},
	}
}

func (p TestParameterSetProvider) Close() error {
	p.TestSecrets.Close()
	return p.TestDocuments.Close()
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

	err = p.ParameterStore.InsertParameterSet(context.Background(), ps)
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
	p.TestSecrets.Create(context.Background(), secrets.SourceSecret, key, value)
}
