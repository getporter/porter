package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkflow(t *testing.T) {
	w := NewWorkflow("dev")

	assert.Equal(t, "dev", w.Namespace, "Namespace was not set")
	assert.NotEmpty(t, w.ID, "ID was not set")
	assert.NotEmpty(t, w.Status.Created, "Created was not set")
	assert.NotEmpty(t, w.Status.Modified, "Modified was not set")
	assert.Equal(t, w.Status.Created, w.Status.Modified, "Created and Modified should have the same timestamp")
	assert.Equal(t, SchemaTypeWorkflow, w.SchemaType, "incorrect SchemaType")
	assert.Equal(t, DefaultWorkflowSchemaVersion, w.SchemaVersion, "incorrect SchemaVersion")
	assert.Empty(t, w.Stages, "Stages should start empty")
}

func TestWorkflow_DefaultDocumentFilter(t *testing.T) {
	w := Workflow{WorkflowSpec: WorkflowSpec{Namespace: "dev"}, ID: "abc123"}
	assert.Equal(t, map[string]any{"namespace": "dev", "_id": "abc123"}, w.DefaultDocumentFilter())
}
