package storage

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/require"
)

var _ WorkflowProvider = TestWorkflowProvider{}

type TestWorkflowProvider struct {
	WorkflowStore
	TestStore
	t         *testing.T
	idCounter uint
}

func NewTestWorkflowProvider(t *testing.T) *TestWorkflowProvider {
	tc := config.NewTestConfig(t)
	testStore := NewTestStore(tc)
	return NewTestWorkflowProviderFor(t, testStore)
}

func NewTestWorkflowProviderFor(t *testing.T, testStore TestStore) *TestWorkflowProvider {
	return &TestWorkflowProvider{
		t:             t,
		TestStore:     testStore,
		WorkflowStore: NewWorkflowStore(testStore),
	}
}

func (p *TestWorkflowProvider) Close() error {
	return p.TestStore.Close()
}

// CreateWorkflow creates a new test workflow and saves it.
func (p *TestWorkflowProvider) CreateWorkflow(w Workflow, transformations ...func(w *Workflow)) Workflow {
	for _, transform := range transformations {
		transform(&w)
	}

	err := p.InsertWorkflow(context.Background(), w)
	require.NoError(p.t, err, "InsertWorkflow failed")
	return w
}

func (p *TestWorkflowProvider) SetMutableWorkflowValues(w *Workflow) {
	p.idCounter += 1
	w.Status.Created = now
	w.Status.Modified = now
}
