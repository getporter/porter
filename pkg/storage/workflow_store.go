package storage

import (
	"context"

	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"get.porter.sh/porter/pkg/tracing"
)

const (
	CollectionWorkflows = "workflows"
)

var _ WorkflowProvider = WorkflowStore{}

// WorkflowStore is a persistent store for workflow documents.
type WorkflowStore struct {
	client BackendClient
}

// NewWorkflowStore creates a persistent store for workflows using the specified
// backing datastore.
func NewWorkflowStore(client BackendClient) WorkflowStore {
	return WorkflowStore{
		client: client,
	}
}

func (s WorkflowStore) InsertWorkflow(ctx context.Context, w Workflow) error {
	w.SchemaVersion = DefaultWorkflowSchemaVersion
	return s.client.put(ctx, protov4.PutRequest_INSERT, w)
}

func (s WorkflowStore) UpsertWorkflow(ctx context.Context, w Workflow) error {
	w.SchemaVersion = DefaultWorkflowSchemaVersion
	return s.client.put(ctx, protov4.PutRequest_UPSERT, w)
}

func (s WorkflowStore) UpdateWorkflow(ctx context.Context, w Workflow) error {
	w.SchemaVersion = DefaultWorkflowSchemaVersion
	return s.client.put(ctx, protov4.PutRequest_UPDATE, w)
}

func (s WorkflowStore) GetWorkflow(ctx context.Context, namespace string, id string) (Workflow, error) {
	return getDocument[Workflow](ctx, s.client, &protov4.GetRequest{
		Kind:      protov4.Kind_KIND_WORKFLOW,
		Namespace: namespace,
		Id:        id,
	})
}

func (s WorkflowStore) ListWorkflows(ctx context.Context, listOptions ListOptions) ([]Workflow, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// Workflows are only filtered by namespace, not by name or labels.
	return listDocuments[Workflow](ctx, s.client, &protov4.ListRequest{
		Kind:      protov4.Kind_KIND_WORKFLOW,
		Namespace: listOptions.Namespace,
		Skip:      listOptions.Skip,
		Limit:     listOptions.Limit,
	})
}

// RemoveWorkflow by its id.
func (s WorkflowStore) RemoveWorkflow(ctx context.Context, namespace string, id string) error {
	_, err := s.client.backend.Remove(ctx, &protov4.RemoveRequest{
		Kind:      protov4.Kind_KIND_WORKFLOW,
		Namespace: namespace,
		Id:        id,
	})
	return err
}
