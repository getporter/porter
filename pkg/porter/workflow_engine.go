package porter

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	depsv2 "get.porter.sh/porter/pkg/cnab/dependencies/v2"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
)

// Engine handles executing a workflow of bundles to execute.
type Engine struct {
	namespace string
	store     storage.InstallationProvider

	// TODO(PEP003): don't inject a resolver, inject the stuff that the resolver uses (store and puller) mock those two instead
	resolver depsv2.BundleGraphResolver
	p        *Porter
}

// NewWorkflowEngine configures a Workflow Engine
func NewWorkflowEngine(namespace string, puller depsv2.BundlePuller, store storage.InstallationProvider, p *Porter) *Engine {
	return &Engine{
		namespace: namespace,
		resolver:  depsv2.NewCompositeResolver(namespace, puller, store),
		store:     store,
		p:         p,
	}
}

// CreateWorkflowOptions are the set of options for creating a Workflow.
type CreateWorkflowOptions struct {
	// DebugMode alters how the workflow is executed so that it can be stepped through.
	DebugMode bool

	// MaxParallel indicates how many parallel bundles may be executed at the same
	// time. Defaults to 0, indicating that the maximum should be determined by the
	// number of available CPUs or cluster nodes (depending on the runtime driver).
	MaxParallel int

	// Installation that triggered the workflow.
	// TODO(PEP003): Does this need to be a full installation? can it just be the spec?
	Installation storage.Installation

	// Bundle definition of the Installation.
	Bundle cnab.ExtendedBundle

	// CustomAction is the name of a custom action defined on the bundle to execute.
	// When not set, the installation is reconciled.
	CustomAction string
}

func (t Engine) CreateWorkflow(ctx context.Context, opts CreateWorkflowOptions) (storage.WorkflowSpec, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	g, err := t.resolver.ResolveDependencyGraph(ctx, opts.Bundle)
	if err != nil {
		return storage.WorkflowSpec{}, err
	}

	nodes, ok := g.Sort()
	if !ok {
		return storage.WorkflowSpec{}, fmt.Errorf("could not generate a workflow for the bundle: the bundle graph has a cyle")
	}

	// Now build job definitions for each node in the graph
	jobs := make(map[string]*storage.Job, len(nodes))
	for _, node := range nodes {
		switch tn := node.(type) {
		case depsv2.BundleNode:
			var inst storage.InstallationSpec
			if tn.IsRoot() {
				inst = opts.Installation.InstallationSpec
			} else {
				// TODO(PEP003?): generate a unique installation name, e.g. ROOTINSTALLATION-DEPKEY-SUFFIX
				// I think we discussed this in a meeting? go look for notes or suggestions
				inst = storage.InstallationSpec{
					Namespace: t.namespace,
					// TODO(PEP003): can we fix the key so that it uses something real from the installation and not root for the root key name?
					Name:   strings.Replace(tn.Key, "root/", opts.Installation.Name+"/", 1),
					Bundle: storage.NewOCIReferenceParts(tn.Reference.Reference),
					// PEP(003): Add labels so that we know who is the parent installation
				}

				// Populate the dependency's credentials from the wiring
				inst.Credentials.SchemaVersion = storage.CredentialSetSchemaVersion
				inst.Credentials.Credentials = make([]secrets.Strategy, 0, len(tn.Credentials))
				for credName, source := range tn.Credentials {
					inst.Credentials.Credentials = append(inst.Credentials.Credentials,
						source.AsWorkflowStrategy(credName, tn.ParentKey))
				}

				// Populate the dependency's parameters from the wiring
				inst.Parameters.SchemaVersion = storage.ParameterSetSchemaVersion
				inst.Parameters.Parameters = make([]secrets.Strategy, 0, len(tn.Parameters))
				for paramName, source := range tn.Parameters {
					inst.Parameters.Parameters = append(inst.Parameters.Parameters,
						source.AsWorkflowStrategy(paramName, tn.ParentKey))
				}
			}

			// Determine which jobs in the workflow we rely upon
			requires := node.GetRequires()
			requiredJobs := make([]string, 0, len(requires))
			for _, requiredKey := range requires {
				// Check if a job was created for this dependency (some won't exist because they are already installed)
				if _, ok := jobs[requiredKey]; ok {
					requiredJobs = append(requiredJobs, requiredKey)
				}
			}

			jobs[tn.Key] = &storage.Job{
				Action:       cnab.ActionInstall, // TODO(PEP003): eventually this needs to support all actions
				Installation: inst,
				Depends:      requiredJobs,
			}

		case depsv2.InstallationNode:
			// TODO(PEP003): Do we need to do anything for this part? Check the status of the installation?

		default:
			return storage.WorkflowSpec{}, fmt.Errorf("invalid node type: %T", tn)
		}

	}

	w := storage.WorkflowSpec{
		Stages: []storage.Stage{
			{Jobs: jobs},
		},
		MaxParallel: opts.MaxParallel,
		DebugMode:   opts.DebugMode,
	}
	return w, nil
}

func (t Engine) RunWorkflow(ctx context.Context, w storage.Workflow) error {
	ctx, span := tracing.StartSpan(ctx, tracing.ObjectAttribute("workflow", w))
	defer span.EndSpan()

	// TODO(PEP003): 2. anything we need to validate?
	w.Prepare()

	// 1. save the workflow to the database
	if err := t.store.UpsertWorkflow(ctx, w); err != nil {
		return err
	}

	// 3. go through each stage and execute it
	for i := range w.Stages {
		// Run each stage in series
		if err := t.executeStage(ctx, w, i); err != nil {
			return fmt.Errorf("stage[%d] failed: %w", i, err)
		}
	}

	/*



		4. what type of status do we want to track? active jobs?
		5. how do we determine which to run? we need to resolve a graph again, with depends. Is there any way to not have to do that?
		6. be smarter
	*/

	return nil
}

func (t Engine) CancelWorkflow(ctx context.Context, workflow storage.Workflow) error {
	//TODO implement me
	// What should cancel do? Mark a status on the record that we check before running the next thing?
	// Who can call this and when?
	panic("implement me")
}

func (t Engine) RetryWorkflow(ctx context.Context, workflow storage.Workflow) error {
	//TODO implement me
	// Start executing from the last failed job (retry the run, keep it and add a second result), skip over everything completed
	panic("implement me")
}

func (t Engine) StepThrough(ctx context.Context, workflow storage.Workflow, job string) error {
	//TODO implement me
	// execute the specified job only and update the status
	panic("implement me")
}

func (t Engine) executeStage(ctx context.Context, w storage.Workflow, stageIndex int) error {
	ctx, span := tracing.StartSpan(ctx, attribute.Int("stage", stageIndex))
	defer span.EndSpan()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s := w.Stages[stageIndex]
	stageGraph := depsv2.NewBundleGraph()
	for _, job := range s.Jobs {
		stageGraph.RegisterNode(job)
	}

	sortedJobs, ok := stageGraph.Sort()
	if !ok {
		return fmt.Errorf("could not sort jobs in stage")
	}

	availableJobs := make(chan *storage.Job, len(s.Jobs))
	completedJobs := make(chan *storage.Job, len(s.Jobs))

	// Default the number of parallel jobs to the number of CPUs
	// This gives us 1 CPU per invocation image.
	var maxParallel int
	if w.DebugMode {
		maxParallel = 1
	} else if w.MaxParallel == 0 {
		maxParallel = runtime.NumCPU()
	} else {
		maxParallel = w.MaxParallel
	}

	// Start up workers to run the jobs as they are available
	jobsInProgress := errgroup.Group{}
	for i := 0; i < maxParallel; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-availableJobs:
					jobsInProgress.Go(func() error {
						// TODO(PEP003) why do we have to look this up again?
						return t.executeJob(ctx, s.Jobs[job.Key], completedJobs)
					})
				}
			}
		}()
	}

	t.queueAvailableJobs(ctx, s, sortedJobs, availableJobs)

	jobsInProgress.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case completedJob := <-completedJobs:
				// If succeeded, remove from the graph so we don't need to keep evaluating it
				// Leave failed ones since they act to stop graph traversal
				if completedJob.Status.IsSucceeded() {
					for i, job := range sortedJobs {
						if job.GetKey() == completedJob.Key {
							if i == 0 {
								sortedJobs = sortedJobs[1:]
							} else {
								sortedJobs = append(sortedJobs[:i-1], sortedJobs[i+1:]...)
							}
							break
						}
					}

					// Look for more jobs to run
					stop := t.queueAvailableJobs(ctx, s, sortedJobs, availableJobs)
					if stop {
						// Stop running all of our jobs
						cancel()
						return nil
					}
				} else {
					return fmt.Errorf("job %s failed: %s", completedJob.Key, completedJob.Status.Message)
				}
			}
		}

	})

	err := jobsInProgress.Wait()
	return err
}

func (t Engine) queueAvailableJobs(ctx context.Context, s storage.Stage, sortedNodes []depsv2.Node, availableJobs chan *storage.Job) bool {
	// Walk through the graph in sorted order (bottom up)
	// if the node's dependencies are all successful, schedule it
	// as soon as it's not schedule, stop looking because none of the remainder will be either
	var i int
	for i = 0; i < len(sortedNodes); i++ {
		node := sortedNodes[i]

		job := node.(*storage.Job)
		switch job.Status.Status {
		case cnab.StatusFailed:
			// stop scheduling more jobs
			return true
		case "":
			jobReady := true
			for _, depKey := range job.Depends {
				dep := s.Jobs[depKey]
				if !dep.Status.IsSucceeded() {
					jobReady = false
					break
				}
			}

			if jobReady {
				availableJobs <- job
				// there are still more jobs to process
				return false
			}
		default:
			continue
		}
	}

	// Did we iterate through all the nodes? Can we stop now?
	return i >= len(sortedNodes)
}

func (t Engine) executeJob(ctx context.Context, j *storage.Job, jobs chan *storage.Job) error {
	ctx, span := tracing.StartSpan(ctx, tracing.ObjectAttribute("job", j))
	defer span.EndSpan()

	opts := ReconcileOptions{
		Installation: j.Installation,
	}
	run, result, err := t.p.ReconcileInstallationInWorkflow(ctx, opts)
	j.Status.LastRunID = run.ID
	j.Status.LastResultID = result.ID
	j.Status.ResultIDs = append(j.Status.ResultIDs, result.ID)
	if err != nil {
		j.Status.Status = cnab.StatusFailed
		j.Status.Message = err.Error()
	} else {
		j.Status.Status = cnab.StatusSucceeded
		j.Status.Message = ""
	}
	jobs <- j
	return nil
}
