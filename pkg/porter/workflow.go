package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/schema"
)

type DisplayWorkflow struct {
	// ID of the workflow.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	SchemaType string `json:"schemaType" yaml:"schemaType"`

	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion"`

	// MaxParallel is the maximum number of jobs that can run in parallel.
	MaxParallel int `json:"maxParallel,omitempty" yaml:"maxParallel,omitempty"`

	// DebugMode tweaks how the workflow is run to make it easier to debug
	DebugMode bool `json:"debugMode,omitempty" yaml:"debugMode,omitempty"`

	// Stages are groups of jobs that run in serial.
	Stages []DisplayStage `json:"stages" yaml:"stages"`

	// TODO(PEP003): When we wrap this in a DisplayWorkflow, override marshal so that we don't marshal an ID or status when empty
	// i.e. if we do a dry run, we shouldn't get out an empty id or status
	Status DisplayWorkflowStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func NewDisplayWorkflow(in storage.Workflow) DisplayWorkflow {
	out := DisplayWorkflow{
		ID:            in.ID,
		SchemaType:    "Workflow",
		SchemaVersion: in.SchemaVersion,
		MaxParallel:   in.MaxParallel,
		DebugMode:     in.DebugMode,
		Stages:        make([]DisplayStage, len(in.Stages)),
		Status:        NewDisplayWorkflowStatus(in.Status),
	}

	for i, inStage := range in.Stages {
		out.Stages[i] = NewDisplayStage(inStage)
	}

	return out
}

// AsSpecOnly creates a subset of user-facing spec fields for exporting to the user
func (w DisplayWorkflow) AsSpecOnly() DisplayWorkflow {
	out := w
	out.ID = ""
	out.Status = DisplayWorkflowStatus{}

	for i, stage := range out.Stages {
		for j, job := range stage.Jobs {
			job = job.AsSpecOnly()
			stage.Jobs[j] = job
		}
		out.Stages[i] = stage
	}
	return out
}

func (w DisplayWorkflow) ToWorkflow() (storage.Workflow, error) {
	out := storage.Workflow{
		ID: cnab.NewULID(),
		WorkflowSpec: storage.WorkflowSpec{
			SchemaType:    storage.SchemaTypeWorkflow,
			SchemaVersion: storage.WorkflowSchemaVersion,
			MaxParallel:   w.MaxParallel,
			DebugMode:     w.DebugMode,
			Stages:        make([]storage.Stage, len(w.Stages)),
		},
	}

	for i, inStage := range w.Stages {
		outStage, err := inStage.ToStage()
		if err != nil {
			return storage.Workflow{}, err
		}
		out.Stages[i] = outStage
	}

	return out, nil
}

type DisplayStage struct {
	Jobs map[string]DisplayJob `json:"jobs" yaml:"jobs"`
}

func (s DisplayStage) ToStage() (storage.Stage, error) {
	out := storage.Stage{
		Jobs: make(map[string]*storage.Job, len(s.Jobs)),
	}

	for jobKey, inJob := range s.Jobs {
		outJob, err := inJob.ToJob(jobKey)
		if err != nil {
			return storage.Stage{}, err
		}
		out.Jobs[jobKey] = outJob
	}
	return out, nil
}

func NewDisplayStage(in storage.Stage) DisplayStage {
	out := DisplayStage{
		Jobs: make(map[string]DisplayJob, len(in.Jobs)),
	}

	for jobKey, inJob := range in.Jobs {
		out.Jobs[jobKey] = NewDisplayJob(*inJob)
	}

	return out
}

type DisplayJob struct {
	Action       string              `json:"action,omitempty" yaml:"action,omitempty"`
	Installation DisplayInstallation `json:"installation" yaml:"installation"`
	Depends      []string            `json:"depends,omitempty" yaml:"depends,omitempty"`
	Status       DisplayJobStatus    `json:"status,omitempty" yaml:"status,omitempty"`
}

func (j DisplayJob) AsSpecOnly() DisplayJob {
	out := j
	out.Installation = out.Installation.AsSpecOnly()
	out.Status = DisplayJobStatus{}
	return out
}

func (j DisplayJob) ToJob(key string) (*storage.Job, error) {

	// We only set the schemaVersion on the workflow and that implies that the schemaVersion for nested types also matches the schema version of related fields.
	// TODO(PEP003): Verify this won't bit us in the 🍑 later
	j.Installation.SchemaVersion = storage.InstallationSchemaVersion

	inst, err := j.Installation.ConvertToInstallation()
	if err != nil {
		return nil, fmt.Errorf("Error converting DisplayInstallation to a storage.Installation: %w", err)
	}

	out := &storage.Job{
		Key:          key,
		Action:       j.Action,
		Installation: inst,
		Depends:      j.Depends,
	}

	return out, nil
}

func NewDisplayJob(in storage.Job) DisplayJob {
	return DisplayJob{
		Action:       in.Action,
		Installation: NewDisplayInstallation(in.Installation),
		Depends:      in.Depends,
		Status:       NewDisplayJobStatus(in.Status),
	}
}

type DisplayJobStatus struct {
	LastRunID    string   `json:"lastRunID" yaml:"lastRunID"`
	LastResultID string   `json:"lastResultID" yaml:"lastResultID"`
	ResultIDs    []string `json:"resultIDs" yaml:"resultIDs"`
	Status       string   `json:"status" yaml:"status"`
	Message      string   `json:"message" yaml:"message"`
}

func NewDisplayJobStatus(in storage.JobStatus) DisplayJobStatus {
	return DisplayJobStatus{
		LastRunID:    in.LastRunID,
		LastResultID: in.LastResultID,
		ResultIDs:    in.ResultIDs,
		Status:       in.Status,
		Message:      in.Message,
	}
}

type DisplayWorkflowStatus struct {
}

func NewDisplayWorkflowStatus(in storage.WorkflowStatus) DisplayWorkflowStatus {
	return DisplayWorkflowStatus{}
}
