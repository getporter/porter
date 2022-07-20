package query

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
)

// MixinQuery allows us to send a command to a bunch of mixins and collect their response.
type MixinQuery struct {
	*portercontext.Context

	// RequireMixinResponse indicates if every mixin must return a successful response.
	// Set to true for required mixin commands that need every mixin to respond to.
	// Set to false for optional mixin commands that every mixin may not have implemented.
	RequireAllMixinResponses bool

	// LogMixinStderr to the supplied contexts Stdout (Context.Out)
	LogMixinErrors bool

	Mixins pkgmgmt.PackageManager
}

// New creates a new instance of a MixinQuery.
func New(cxt *portercontext.Context, mixins pkgmgmt.PackageManager) *MixinQuery {
	return &MixinQuery{
		Context: cxt,
		Mixins:  mixins,
	}
}

// MixinInputGenerator provides data about the mixins to the MixinQuery
// for it to execute upon.
type MixinInputGenerator interface {
	// ListMixins provides the list of mixin names to query over.
	ListMixins() []string

	// BuildInput generates the input to send to the specified mixin given its name.
	BuildInput(mixinName string) ([]byte, error)
}

// Execute the specified command using an input generator.
// For example, the ManifestGenerator will iterate over the mixins in a manifest and send
// them their config and the steps associated with their mixin.
func (q *MixinQuery) Execute(ctx context.Context, cmd string, inputGenerator MixinInputGenerator) (map[string]string, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	mixinNames := inputGenerator.ListMixins()
	results := make(map[string]string, len(mixinNames))
	type queryResponse struct {
		mixinName string
		output    string
		runErr    error
	}

	var responses = make(chan queryResponse, len(mixinNames))
	gerr := errgroup.Group{}

	for _, mn := range mixinNames {
		mixinName := mn // Force mixinName to be in the go routine's closure below
		gerr.Go(func() error {
			// Copy the existing context and tweak to pipe the output differently
			mixinStdout := &bytes.Buffer{}
			mixinContext := *q.Context
			mixinContext.Out = mixinStdout // mixin stdout -> mixin response

			if q.LogMixinErrors {
				mixinContext.Err = q.Context.Out // mixin stderr -> porter logs
			} else {
				mixinContext.Err = ioutil.Discard
			}

			inputB, err := inputGenerator.BuildInput(mixinName)
			if err != nil {
				return err
			}

			cmd := pkgmgmt.CommandOptions{
				Command: cmd,
				Input:   string(inputB),
			}
			runErr := q.Mixins.Run(ctx, &mixinContext, mixinName, cmd)

			// Pack the error from running the command in the response so we can
			// decide if we care about it, if we returned it normally, the
			// waitgroup will short circuit immediately on the first error
			responses <- queryResponse{
				mixinName: mixinName,
				output:    mixinStdout.String(),
				runErr:    runErr,
			}
			return nil
		})
	}

	err := gerr.Wait()
	close(responses)
	if err != nil {
		return nil, err
	}

	// Collect responses and errors
	var runErr error
	for response := range responses {
		if response.runErr == nil {
			results[response.mixinName] = response.output
		} else {
			runErr = multierror.Append(runErr,
				fmt.Errorf("error encountered from mixin %q: %w", response.mixinName, response.runErr))
		}
	}

	if runErr != nil {
		if q.RequireAllMixinResponses {
			return nil, span.Error(runErr)
		}

		// This is a debug because we expect not all mixins to implement some
		// optional commands, like lint and don't want to print their error
		// message when we query them with a command they don't support.
		span.Debugf(fmt.Errorf("not all mixins responded successfully: %w", runErr).Error())
	}

	return results, nil
}
