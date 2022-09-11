package porter

import (
	"context"
	"fmt"
	"regexp"

	"get.porter.sh/porter/pkg/storage"

	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"get.porter.sh/porter/pkg/tracing"
)

// PorterSecretStrategy knows how to resolve specially formatted wiring strings
// such as workflow.jobs.db.outputs.connstr from Porter instead of from a plugin.
// It is not written as a plugin because it is much more straightforward to
// retrieve the data already loaded in the running Porter instance than to start
// another one, load its config and requery the database.
type PorterSecretStrategy struct {
	installations storage.InstallationProvider
}

// regular expression for parsing a workflow wiring string, such as workflow.jobs.db.outputs.connstr
var workflowWiringRegex = regexp.MustCompile(`workflow\.jobs\.([^\.]+)\.(.+)`)

func (s PorterSecretStrategy) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// TODO(PEP003): It would be great when we configure this strategy that we also do host, so that host secret resolution isn't deferred to the plugins
	// i.e. we can configure a secret strategy and still be able to resolve directly in porter any host values.
	if keyName != "porter" {
		return "", fmt.Errorf("attempted to resolve secrets of type %s from the porter strategy", keyName)
	}

	wiring, err := v2.ParseWorkflowWiring(keyValue)
	if err != nil {
		return "", fmt.Errorf("invalid workflow wiring was passed to the porter strategy, %s", keyValue)
	}

	// TODO(PEP003): How do we want to re-resolve credentials passed to the root bundle? They aren't recorded so it's not a simple lookup
	if wiring.Parameter != "" {
		// TODO(PEP003): Resolve a parameter from another job that has not run yet
		// 1. Find the workflow definition from the db (need a way to track "current" workflow)
		// 2. Grab the job based on the jobid in the workflow wiring
		// 3. First check the parameters field for the param, resolve just that if available, otherwise resolve parameter sets and get it from there
		// it sure would help if we remembered what params are in each set
	} else if wiring.Output != "" {
		// TODO(PEP003): Resolve the output from an already executed job
	}

	panic("not implemented")
}
