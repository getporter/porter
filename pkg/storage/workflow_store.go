package storage

import (
	"context"

	"get.porter.sh/porter/pkg/tracing"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	CollectionWorkflows = "workflows"
)

var _ WorkflowProvider = WorkflowStore{}

// WorkflowStore is a persistent store for workflow documents.
type WorkflowStore struct {
	store Store
}

// NewWorkflowStore creates a persistent store for workflows using the specified
// backing datastore.
func NewWorkflowStore(datastore Store) WorkflowStore {
	return WorkflowStore{
		store: datastore,
	}
}

// EnsureWorkflowIndices creates indices on the workflows collection.
func EnsureWorkflowIndices(ctx context.Context, store Store) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debug("Initializing workflow collection indices")

	opts := EnsureIndexOptions{
		Indices: []Index{
			// query workflows by namespace (list) or namespace + id (get)
			{Collection: CollectionWorkflows, Keys: []string{"namespace", "_id"}, Unique: true},
		},
	}

	err := store.EnsureIndex(ctx, opts)
	return span.Error(err)
}

func (s WorkflowStore) InsertWorkflow(ctx context.Context, w Workflow) error {
	w.SchemaVersion = DefaultWorkflowSchemaVersion
	opts := InsertOptions{
		Documents: []any{w},
	}
	return s.store.Insert(ctx, CollectionWorkflows, opts)
}

func (s WorkflowStore) UpsertWorkflow(ctx context.Context, w Workflow) error {
	w.SchemaVersion = DefaultWorkflowSchemaVersion
	opts := UpdateOptions{
		Upsert:   true,
		Document: w,
	}
	return s.store.Update(ctx, CollectionWorkflows, opts)
}

func (s WorkflowStore) UpdateWorkflow(ctx context.Context, w Workflow) error {
	w.SchemaVersion = DefaultWorkflowSchemaVersion
	opts := UpdateOptions{
		Document: w,
	}
	return s.store.Update(ctx, CollectionWorkflows, opts)
}

func (s WorkflowStore) GetWorkflow(ctx context.Context, namespace string, id string) (Workflow, error) {
	var out Workflow

	opts := FindOptions{
		Filter: bson.M{
			"namespace": namespace,
			"_id":       id,
		},
	}

	err := s.store.FindOne(ctx, CollectionWorkflows, opts, &out)
	return out, err
}

func (s WorkflowStore) ListWorkflows(ctx context.Context, listOptions ListOptions) ([]Workflow, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	filter := bson.M{}
	if listOptions.Namespace != "*" {
		filter["namespace"] = listOptions.Namespace
	}

	opts := FindOptions{
		Sort:   []string{"namespace", "_id"},
		Filter: filter,
		Skip:   listOptions.Skip,
		Limit:  listOptions.Limit,
	}

	var out []Workflow
	err := s.store.Find(ctx, CollectionWorkflows, opts, &out)
	return out, err
}

// RemoveWorkflow by its id.
func (s WorkflowStore) RemoveWorkflow(ctx context.Context, namespace string, id string) error {
	opts := RemoveOptions{
		Filter: bson.M{
			"namespace": namespace,
			"_id":       id,
		},
	}
	return s.store.Remove(ctx, CollectionWorkflows, opts)
}
