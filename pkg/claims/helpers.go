package claims

import (
	"context"
	"fmt"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/require"
)

var (
	_ Provider = TestClaimProvider{}

	// A fixed now timestamp that we can use for comparisons in tests
	now = time.Date(2020, time.April, 18, 1, 2, 3, 4, time.UTC)

	installationID = "01FZVC5AVP8Z7A78CSCP1EJ604"
)

type TestClaimProvider struct {
	ClaimStore
	storage.TestStore
	t         *testing.T
	idCounter uint
}

func NewTestClaimProvider(t *testing.T) *TestClaimProvider {
	tc := config.NewTestConfig(t)
	testStore := storage.NewTestStore(tc)
	return NewTestClaimProviderFor(t, testStore)
}

func NewTestClaimProviderFor(t *testing.T, testStore storage.TestStore) *TestClaimProvider {
	return &TestClaimProvider{
		t:          t,
		TestStore:  testStore,
		ClaimStore: NewClaimStore(testStore),
	}
}

func (p *TestClaimProvider) Close() error {
	return p.TestStore.Close()
}

// CreateInstallation creates a new test installation and saves it.
func (p *TestClaimProvider) CreateInstallation(i Installation, transformations ...func(i *Installation)) Installation {
	for _, transform := range transformations {
		transform(&i)
	}

	err := p.InsertInstallation(context.Background(), i)
	require.NoError(p.t, err, "InsertInstallation failed")
	return i
}

func (p *TestClaimProvider) SetMutableInstallationValues(i *Installation) {
	i.ID = installationID
	i.Status.Created = now
	i.Status.Modified = now
}

// CreateRun creates a new claim and saves it.
func (p *TestClaimProvider) CreateRun(r Run, transformations ...func(r *Run)) Run {
	for _, transform := range transformations {
		transform(&r)
	}

	err := p.InsertRun(context.Background(), r)
	require.NoError(p.t, err, "InsertRun failed")
	return r
}

func (p *TestClaimProvider) SetMutableRunValues(r *Run) {
	p.idCounter += 1
	r.ID = fmt.Sprintf("%d", p.idCounter)
	r.Created = now
}

// CreateResult creates a new result from the specified claim and saves it.
func (p *TestClaimProvider) CreateResult(r Result, transformations ...func(r *Result)) Result {
	for _, transform := range transformations {
		transform(&r)
	}

	err := p.InsertResult(context.Background(), r)
	require.NoError(p.t, err, "InsertResult failed")
	return r
}

func (p *TestClaimProvider) SetMutableResultValues(r *Result) {
	p.idCounter += 1
	r.ID = fmt.Sprintf("%d", p.idCounter)
	r.Created = now
}

// CreateOutput creates a new output from the specified claim and result and saves it.
func (p *TestClaimProvider) CreateOutput(o Output, transformations ...func(o *Output)) Output {
	for _, transform := range transformations {
		transform(&o)
	}

	err := p.InsertOutput(context.Background(), o)
	require.NoError(p.t, err, "InsertOutput failed")
	return o
}
