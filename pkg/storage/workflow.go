package storage

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/cnabio/cnab-go/schema"
)

// Workflow represents how a bundle and its dependencies should be run by Porter.
type Workflow struct {
	// ID of the workflow.
	ID string `json:"id"`

	WorkflowSpec `json:"spec"`

	// TODO(PEP003): When we wrap this in a DisplayWorkflow, override marshal so that we don't marshal an ID or status when empty
	// i.e. if we do a dry run, we shouldn't get out an empty id or status
	Status WorkflowStatus `json:"status"`
}

type WorkflowSpec struct {
	SchemaVersion schema.Version `json:"schemaVersion"`

	// MaxParallel is the maximum number of jobs that can run in parallel.
	MaxParallel int `json:"maxParallel"`

	// DebugMode tweaks how the workflow is run to make it easier to debug
	DebugMode bool `json:"debugMode"`

	// Stages are groups of jobs that run in serial.
	Stages []Stage `json:"stages"`
}

// GetJob finds the specified job by its key.
// All job keys within a workflow must be unique.
func (w *WorkflowSpec) GetJob(jobKey string) (*Job, error) {
	for _, s := range w.Stages {
		for k, j := range s.Jobs {
			if k == jobKey {
				return j, nil
			}
		}
	}

	return nil, fmt.Errorf("workflow does not contain job key %s", jobKey)
}

// TODO(PEP003): Figure out what needs to be persisted, and how to persist multiple or continued runs
type WorkflowStatus struct {
}

// Prepare updates the internal data representation of the workflow before running it.
func (w *Workflow) Prepare() {
	// Assign an id to the workflow if needed
	if w.ID == "" {
		w.ID = cnab.NewULID()
	}

	for _, s := range w.Stages {
		s.Prepare(w.ID)
	}
}

// Stage represents a set of jobs that should run, possibly in parallel.
type Stage struct {
	// Jobs is the set of bundles to execute, keyed by the job name.
	Jobs map[string]*Job `json:"jobs"`
}

func (s *Stage) Prepare(workflowID string) {
	// Update the jobs so that they know their job key (since they won't be used within the larger workflow, but as independent jobs)
	for jobKey, job := range s.Jobs {
		job.Prepare(workflowID, jobKey)
	}
}

// Job represents the execution of a bundle.
type Job struct {
	// TODO(PEP003): no job can have the same name as the parent installation (which is keyed from the installation). Or we need to stick to root and reserve that?
	// Key is the unique name of the job within a stage.
	// We handle copying this value so that it's easier to work with a single job when not in a map
	Key string `json:"-"`

	// Action name to execute on the bundle, when empty default to applying the installation.
	Action string `json:"action"`

	// TODO(PEP003): workflows should have DisplayWorkflow and use DisplayInstallation
	// Installation defines the installation upon which Porter should act.
	Installation InstallationSpec `json:"installation"`

	// Depends is a list of job keys that the Job depends upon.
	Depends []string `json:"depends"`

	Status JobStatus `json:"status,omitempty"`
}

func (j *Job) GetRequires() []string {
	return j.Depends
}

func (j *Job) GetKey() string {
	return j.Key
}

func (j *Job) Prepare(workflowId string, jobKey string) {
	j.Key = jobKey
	for i, param := range j.Installation.Parameters.Parameters {
		if param.Source.Key != "porter" {
			continue
		}

		// Update the template string with the id of the current workflow
		param.Source.Value = strings.Replace(param.Source.Value, "workflow.jobs", fmt.Sprintf("workflow.%s.jobs", workflowId), 1)
		j.Installation.Parameters.Parameters[i] = param
	}
}

type JobStatus struct {
	LastRunID    string   `json:lastRunID`
	LastResultID string   `json:lastResultID`
	ResultIDs    []string `json:resultIDs`
	Status       string   `json:"status"`
	Message      string   `json:"message"`
}

func (s JobStatus) IsSucceeded() bool {
	return s.Status == cnab.StatusSucceeded
}

func (s JobStatus) IsFailed() bool {
	return s.Status == cnab.StatusFailed
}

func (s JobStatus) IsDone() bool {
	switch s.Status {
	case cnab.StatusSucceeded, cnab.StatusFailed:
		return true
	}
	return false
}
