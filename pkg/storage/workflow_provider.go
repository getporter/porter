package storage

import (
	"context"
)

// WorkflowProvider is an interface for interacting with Porter's workflow data.
type WorkflowProvider interface {
	// InsertWorkflow saves a new Workflow document.
	InsertWorkflow(ctx context.Context, w Workflow) error

	// UpsertWorkflow saves a Workflow document, creating it if it doesn't already exist.
	UpsertWorkflow(ctx context.Context, w Workflow) error

	// UpdateWorkflow saves changes to an existing Workflow document.
	UpdateWorkflow(ctx context.Context, w Workflow) error

	// GetWorkflow retrieves a Workflow document by namespace and id.
	GetWorkflow(ctx context.Context, namespace string, id string) (Workflow, error)

	// ListWorkflows returns Workflows sorted in ascending order by namespace and then id.
	ListWorkflows(ctx context.Context, listOptions ListOptions) ([]Workflow, error)

	// RemoveWorkflow by its id.
	RemoveWorkflow(ctx context.Context, namespace string, id string) error
}
