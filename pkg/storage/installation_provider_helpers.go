package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/require"
)

var (
	_ InstallationProvider = TestInstallationProvider{}

	// A fixed now timestamp that we can use for comparisons in tests
	now = time.Date(2020, time.April, 18, 1, 2, 3, 4, time.UTC)

	installationID = "01FZVC5AVP8Z7A78CSCP1EJ604"
)

type TestInstallationProvider struct {
	InstallationStore
	TestStore
	t         *testing.T
	idCounter uint
}

func NewTestInstallationProvider(t *testing.T) *TestInstallationProvider {
	tc := config.NewTestConfig(t)
	testStore := NewTestStore(tc)
	return NewTestInstallationProviderFor(t, testStore)
}

func NewTestInstallationProviderFor(t *testing.T, testStore TestStore) *TestInstallationProvider {
	return &TestInstallationProvider{
		t:                 t,
		TestStore:         testStore,
		InstallationStore: NewInstallationStore(testStore),
	}
}

func (p *TestInstallationProvider) Close() error {
	return p.TestStore.Close()
}

// CreateInstallation creates a new test installation and saves it.
func (p *TestInstallationProvider) CreateInstallation(i Installation, transformations ...func(i *Installation)) Installation {
	for _, transform := range transformations {
		transform(&i)
	}

	err := p.InsertInstallation(context.Background(), i)
	require.NoError(p.t, err, "InsertInstallation failed")
	return i
}

func (p *TestInstallationProvider) SetMutableInstallationValues(i *Installation) {
	i.ID = installationID
	i.Status.Created = now
	i.Status.Modified = now
}

// CreateRun creates a new test run and saves it.
func (p *TestInstallationProvider) CreateRun(r Run, transformations ...func(r *Run)) Run {
	for _, transform := range transformations {
		transform(&r)
	}

	err := p.InsertRun(context.Background(), r)
	require.NoError(p.t, err, "InsertRun failed")
	return r
}

func (p *TestInstallationProvider) SetMutableRunValues(r *Run) {
	p.idCounter += 1
	r.ID = fmt.Sprintf("%d", p.idCounter)
	r.Created = now
}

// CreateResult creates a new test result and saves it.
func (p *TestInstallationProvider) CreateResult(r Result, transformations ...func(r *Result)) Result {
	for _, transform := range transformations {
		transform(&r)
	}

	err := p.InsertResult(context.Background(), r)
	require.NoError(p.t, err, "InsertResult failed")
	return r
}

func (p *TestInstallationProvider) SetMutableResultValues(r *Result) {
	p.idCounter += 1
	r.ID = fmt.Sprintf("%d", p.idCounter)
	r.Created = now
}

// CreateOutput creates a new test output and saves it.
func (p *TestInstallationProvider) CreateOutput(o Output, transformations ...func(o *Output)) Output {
	for _, transform := range transformations {
		transform(&o)
	}

	err := p.InsertOutput(context.Background(), o)
	require.NoError(p.t, err, "InsertOutput failed")
	return o
}
