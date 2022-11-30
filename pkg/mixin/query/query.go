package query

import (
	"bytes"
	"context"
	"fmt"
	"io"

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

type MixinBuildOutput struct {
	// Name of the mixin.
	Name string

	// Stdout is the contents of stdout from calling the mixin.
	Stdout string

	// Error returned when the mixin was called.
	Error error
}

// Execute the specified command using an input generator.
// For example, the ManifestGenerator will iterate over the mixins in a manifest and send
// them their config and the steps associated with their mixin.
// The mixins are queried in parallel but the results are sorted in the order that the mixins were defined in the manifest.
func (q *MixinQuery) Execute(ctx context.Context, cmd string, inputGenerator MixinInputGenerator) ([]MixinBuildOutput, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	mixinNames := inputGenerator.ListMixins()
	results := make([]MixinBuildOutput, len(mixinNames))
	gerr := errgroup.Group{}

	for i, mn := range mixinNames {
		// Force variables to be in the go routine's closure below
		i := i
		mn := mn
		results[i].Name = mn

		gerr.Go(func() error {
			// Copy the existing context and tweak to pipe the output differently
			mixinStdout := &bytes.Buffer{}
			mixinContext := *q.Context
			mixinContext.Out = mixinStdout // mixin stdout -> mixin result

			if q.LogMixinErrors {
				mixinContext.Err = q.Context.Out // mixin stderr -> porter logs
			} else {
				mixinContext.Err = io.Discard
			}

			inputB, err := inputGenerator.BuildInput(mn)
			if err != nil {
				return err
			}

			cmd := pkgmgmt.CommandOptions{
				Command: cmd,
				Input:   string(inputB),
			}
			runErr := q.Mixins.Run(ctx, &mixinContext, mn, cmd)

			results[i].Stdout = mixinStdout.String()
			results[i].Error = runErr
			return nil
		})
	}

	err := gerr.Wait()
	if err != nil {
		return nil, err
	}

	// Collect responses and errors
	var runErr error
	for _, result := range results {
		if result.Error != nil {
			runErr = multierror.Append(runErr,
				fmt.Errorf("error encountered from mixin %q: %w", result.Name, result.Error))
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
