package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ WorkflowProvider = WorkflowStore{}

func TestWorkflowStore_InsertGetListUpdateRemove(t *testing.T) {
	cp := NewTestWorkflowProvider(t)
	defer cp.Close()

	w := cp.CreateWorkflow(NewWorkflow("dev"), func(w *Workflow) {
		w.ID = "workflow-1"
		w.Root = "root"
		w.Stages = []Stage{
			{Jobs: map[string]Job{
				"root": {ID: "root", Action: "install", Installation: NewInstallation("dev", "foo")},
			}},
		}
	})

	t.Run("get", func(t *testing.T) {
		got, err := cp.GetWorkflow(context.Background(), "dev", "workflow-1")
		require.NoError(t, err)
		assert.Equal(t, w.ID, got.ID)
		assert.Equal(t, w.Root, got.Root)
		require.Len(t, got.Stages, 1)
		assert.Contains(t, got.Stages[0].Jobs, "root")
	})

	t.Run("list", func(t *testing.T) {
		list, err := cp.ListWorkflows(context.Background(), ListOptions{Namespace: "dev"})
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("update", func(t *testing.T) {
		w.Status.Status = "running"
		err := cp.UpdateWorkflow(context.Background(), w)
		require.NoError(t, err)

		got, err := cp.GetWorkflow(context.Background(), "dev", "workflow-1")
		require.NoError(t, err)
		assert.Equal(t, "running", got.Status.Status)
	})

	t.Run("remove", func(t *testing.T) {
		err := cp.RemoveWorkflow(context.Background(), "dev", "workflow-1")
		require.NoError(t, err)

		_, err = cp.GetWorkflow(context.Background(), "dev", "workflow-1")
		require.Error(t, err)
	})
}
