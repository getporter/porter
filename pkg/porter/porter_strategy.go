package porter

import (
	"context"

	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"go.mongodb.org/mongo-driver/bson"
)

// PorterSecretStrategy knows how to resolve specially formatted wiring strings
// such as workflow.jobs.db.outputs.connstr from Porter instead of from a plugin.
// It is not written as a plugin because it is much more straightforward to
// retrieve the data already loaded in the running Porter instance than to start
// another one, load its config and requery the database.
// This should always run in-process within porter and never as an out-of-process plugin.
type PorterSecretStrategy struct {
	porter *Porter
}

func NewPorterSecretStrategy(p *Porter) PorterSecretStrategy {
	return PorterSecretStrategy{porter: p}
}

func (s PorterSecretStrategy) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// TODO(PEP003): It would be great when we configure this strategy that we also do host, so that host secret resolution isn't deferred to the plugins
	// i.e. we can configure a secret strategy and still be able to resolve directly in porter any host values.
	if keyName != "porter" {
		return "", span.Errorf("attempted to resolve secrets of type %s from the porter strategy", keyName)
	}

	wiring, err := v2.ParseWorkflowWiring(keyValue)
	if err != nil {
		return "", span.Errorf("invalid workflow wiring was passed to the porter strategy, %s", keyValue)
	}

	// We support retrieving certain data from Porter's database:
	// workflow.WORKFLOWID.jobs.JOBKEY.outputs.OUTPUT
	// The WORKFLOWID is set to the current executing workflow by the workflow engine before running the job
	// It is not stored in the database and is always set dynamically when the job is run.

	w, err := s.porter.Installations.GetWorkflow(ctx, wiring.WorkflowID)
	if err != nil {
		return "", span.Errorf("error retrieving workflow %s: %w", wiring.WorkflowID, err)
	}

	// Prepare internal data structures of the workflow
	w.Prepare()

	// locate the job in the workflow
	j, err := w.GetJob(wiring.JobKey)
	if err != nil {
		return "", span.Errorf("error retrieving job from workflow %s: %w", wiring.WorkflowID)
	}

	if j.Status.LastResultID == "" {
		return "", span.Errorf("error retrieving job status for %s in workflow %s, no result recorded yet", wiring.JobKey, wiring.WorkflowID)
	}

	/*

	 */

	// TODO(PEP003): How do we want to re-resolve credentials passed to the root bundle? They aren't recorded so it's not a simple lookup
	if wiring.Credential != "" {

	}
	else if wiring.Parameter != "" {
		// TODO(PEP003): Resolve a parameter from another job that has not run yet
		// IS THIS ACTUALLY A PROBLEM? We pass creds/params from the root job, which we need to deal with, but otherwise we only pass outputs from non-root jobs
		// 1. Find the workflow definition from the db (need a way to track "current" workflow)
		// 2. Grab the job based on the jobid in the workflow wiring
		// 3. First check the parameters field for the param, resolve just that if available, otherwise resolve parameter sets and get it from there
		// it sure would help if we remembered what params are in each set

		// TODO(PEP003): For now resolve all params, but in the future resolve the parameters on an installation but filter which params you care about so that you don't resolve stuff that isn't used

		return "", nil
	} else if wiring.Output != "" {
		// Lookup the result and run associated with the job run in that workflow
		outputs, err := s.porter.Installations.FindOutputs(ctx, storage.FindOptions{
			Sort:  []string{"-_id"},
			Skip:  0,
			Limit: 1,
			Filter: bson.M{
				"resultId": j.Status.LastResultID,
				"name":     wiring.Output,
			},
		})
		if err != nil {
			// TODO(PEP003): Move a lot of these values into the span attributes instead of in the error message
			return "", span.Errorf("error retrieving output %s from result %s for job %s in workflow %s: %w", wiring.Output, j.Status.LastResultID, wiring.JobKey, wiring.WorkflowID)
		}

		if len(outputs) == 0 {
			return "", span.Errorf("no output named %s, found for result %s for job %s in workflow %s", wiring.Output, j.Status.LastResultID, wiring.JobKey, wiring.WorkflowID)
		}

		output, err := s.porter.Sanitizer.RestoreOutput(ctx, outputs[1])
		if err != nil {
			return "", span.Errorf("error restoring output named %s, found for result %s for job %s in workflow %s", wiring.Output, j.Status.LastResultID, wiring.JobKey, wiring.WorkflowID)
		}

		return string(output.Value), nil
	}

	panic("not implemented")
}
