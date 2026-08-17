package storage

import (
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
)

var _ Document = Workflow{}

// Workflow is a request to execute a set of bundles: a root bundle and its
// resolved dependency graph. Like Run, it represents an execution-request
// instance (keyed by ID) rather than a named singleton resource like
// Installation.
//
// A Workflow only records what to run and in what order; it does not resolve
// or store parameter/credential values. Job.Installation.Parameters and
// Job.Credentials may reference another job's not-yet-produced output (via
// the "porter" secrets strategy, hint format
// workflow.jobs.<jobID>.outputs.<name>, see
// v2.DependencySource.AsWorkflowWiring), but those references are only
// resolved just-in-time when the ExecutionPlan generated from this workflow
// is run, matching how Run.Parameters/Run.Credentials are resolved JIT
// before execution.
type Workflow struct {
	// ID of the Workflow.
	ID string `json:"_id"`

	WorkflowSpec

	// Status of the workflow.
	Status WorkflowStatus `json:"status,omitzero"`
}

// WorkflowSpec contains workflow fields that represent the desired state of the workflow.
type WorkflowSpec struct {
	// SchemaType indicates the type of resource.
	SchemaType string `json:"schemaType"`

	// SchemaVersion is the version of the workflow state schema.
	SchemaVersion cnab.SchemaVersion `json:"schemaVersion"`

	// Namespace in which the workflow is defined.
	Namespace string `json:"namespace"`

	// Root is the ID of the Job representing the bundle that was directly
	// requested to run, as opposed to a Job created to satisfy a dependency.
	Root string `json:"root"`

	// MaxParallel is the maximum number of jobs within a stage that may run
	// in parallel. Zero means the number should be determined at run time,
	// for example by the number of available CPUs.
	MaxParallel int `json:"maxParallel,omitempty"`

	// DebugMode runs jobs one at a time, even within a stage, to make the
	// workflow easier to step through.
	DebugMode bool `json:"debugMode,omitempty"`

	// Stages are groups of jobs that run in series; within a stage, jobs may
	// run in parallel (subject to MaxParallel) once their DependsOn jobs have
	// completed.
	Stages []Stage `json:"stages,omitempty"`
}

// Stage is a group of jobs that may run in parallel, once each job's
// DependsOn jobs have completed. Stages themselves run in series.
type Stage struct {
	// Jobs in this stage, keyed by job ID. This addressing scheme is already
	// assumed by v2.DependencySource.AsWorkflowWiring, which produces
	// references of the form workflow.jobs.<jobID>.outputs.<name>.
	Jobs map[string]Job `json:"jobs"`
}

// Job is one installation action within a workflow, for example installing
// or upgrading a specific installation with a specific bundle reference. It
// is the persistable form of a Node in a resolved dependency graph.
type Job struct {
	// ID of the job, unique within the workflow.
	ID string `json:"id"`

	// Alias is the dependency name from the parent bundle's `requires` map
	// that this job was created to satisfy. Empty for the root job.
	Alias string `json:"alias,omitempty"`

	// Action to execute against the installation, e.g. install, upgrade, uninstall.
	Action string `json:"action"`

	// Installation is the full desired state of the installation this job
	// acts on. A dependency's installation may not exist yet and must be
	// created, so the job carries everything needed to do that (bundle
	// reference, labels, credential/parameter set names), not just a
	// namespace/name reference.
	//
	// Parameter wiring (a value sourced from another job's not-yet-produced
	// output) is represented as an entry in Installation.Parameters.Parameters
	// with Source.Strategy "porter" and Source.Hint set to the workflow
	// wiring reference, reusing the same override mechanism already used for
	// user-supplied parameter overrides.
	Installation Installation `json:"installation"`

	// Credentials to resolve for this job. Installation has no field for
	// inline credential values (only named CredentialSets), so credential
	// wiring is recorded here instead, using the same SourceMap/Source
	// mechanism as Installation.Parameters.Parameters: Source.Strategy
	// "porter" and Source.Hint set to the workflow wiring reference.
	Credentials secrets.StrategyList `json:"credentials,omitempty"`

	// DependsOn is the set of job IDs that must complete before this job can
	// run. Derived from both structural (requires) and data-flow (wiring)
	// edges in the dependency graph this workflow was built from.
	DependsOn []string `json:"dependsOn,omitempty"`

	// SharingGroup is the dependency's sharing group name, if any.
	SharingGroup string `json:"sharingGroup,omitempty"`
}

// WorkflowStatus contains status fields for a workflow that are set as
// its jobs are run.
type WorkflowStatus struct {
	// Created timestamp of the workflow.
	Created time.Time `json:"created"`

	// Modified timestamp of the workflow.
	Modified time.Time `json:"modified"`

	// Status of the workflow overall, using the same status values as
	// Result.Status, e.g. cnab.StatusPending, cnab.StatusRunning.
	Status string `json:"status,omitempty"`

	// Jobs mirrors the workflow's jobs by ID, caching each job's run
	// history and latest outcome, the same way InstallationStatus caches an
	// Installation's last run.
	Jobs map[string]JobStatus `json:"jobs,omitempty"`
}

// JobStatus is a cache of a job's run history and latest outcome.
type JobStatus struct {
	// RunID of the job's most recent run.
	RunID string `json:"runId,omitempty"`

	// ResultID of the job's most recent run.
	ResultID string `json:"resultId,omitempty"`

	// ResultIDs is every result recorded for this job, in order, allowing a
	// retried job to keep its full history.
	ResultIDs []string `json:"resultIds,omitempty"`

	// Status of the job's most recent run, e.g. cnab.StatusPending, cnab.StatusRunning.
	Status string `json:"status,omitempty"`

	// Message communicates the outcome of the job's most recent run, for
	// example an error message when Status is cnab.StatusFailed.
	Message string `json:"message,omitempty"`
}

func (w Workflow) DefaultDocumentFilter() map[string]any {
	return map[string]any{"_id": w.ID}
}

// NewWorkflow creates a workflow document with default values initialized.
func NewWorkflow(namespace string) Workflow {
	now := time.Now()
	return Workflow{
		ID: cnab.NewULID(),
		WorkflowSpec: WorkflowSpec{
			SchemaType:    SchemaTypeWorkflow,
			SchemaVersion: DefaultWorkflowSchemaVersion,
			Namespace:     namespace,
		},
		Status: WorkflowStatus{
			Created:  now,
			Modified: now,
		},
	}
}
