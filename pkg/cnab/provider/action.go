package cnabprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	cnabaction "github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/driver"
	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"
)

// Shared arguments for all CNAB actions
type ActionArguments struct {
	// Action to execute, e.g. install, upgrade.
	Action string

	// Name of the installation.
	Installation storage.Installation

	// BundleReference is the set of information necessary to execute a bundle.
	BundleReference cnab.BundleReference

	// Additional files to copy into the bundle
	// Target Path => File Contents
	Files map[string]string

	// Params is the fully resolved set of parameters.
	Params map[string]interface{}

	// Driver is the CNAB-compliant driver used to run bundle actions.
	Driver string

	// Give the bundle privileged access to the docker daemon.
	AllowDockerHostAccess bool

	// PersistLogs specifies if the invocation image output should be saved as an output.
	PersistLogs bool
}

func (r *Runtime) ApplyConfig(ctx context.Context, args ActionArguments) cnabaction.OperationConfigs {
	return cnabaction.OperationConfigs{
		r.SetOutput(),
		r.AddFiles(ctx, args),
		r.AddEnvironment(args),
		r.AddRelocation(args),
	}
}

func (r *Runtime) SetOutput() cnabaction.OperationConfigFunc {
	return func(op *driver.Operation) error {
		op.Out = r.Out
		op.Err = r.Err
		return nil
	}
}

func (r *Runtime) AddFiles(ctx context.Context, args ActionArguments) cnabaction.OperationConfigFunc {
	return func(op *driver.Operation) error {
		for k, v := range args.Files {
			op.Files[k] = v
		}

		// Add claim.json to file list as well, if exists
		claim, err := r.installations.GetLastRun(ctx, args.Installation.Namespace, args.Installation.Name)
		if err == nil {
			claimBytes, err := json.Marshal(claim)
			if err != nil {
				return fmt.Errorf("could not marshal claim %s for installation %s: %w", claim.ID, args.Installation, err)
			}
			op.Files[config.ClaimFilepath] = string(claimBytes)
		}

		return nil
	}
}

func (r *Runtime) AddEnvironment(args ActionArguments) cnabaction.OperationConfigFunc {
	const verbosityEnv = "PORTER_VERBOSITY"

	return func(op *driver.Operation) error {
		op.Environment[config.EnvPorterInstallationNamespace] = args.Installation.Namespace
		op.Environment[config.EnvPorterInstallationName] = args.Installation.Name

		// Pass the verbosity from porter's local config into the bundle
		op.Environment[verbosityEnv] = r.Config.GetVerbosity().Level().String()

		// When a bundle is run in debug mode, the verbosity is automatically set to debug
		if debugMode, _ := args.Params["porter-debug"].(bool); debugMode {
			op.Environment[verbosityEnv] = zapcore.DebugLevel.String()
		}
		return nil
	}
}

// AddRelocation operates on an ActionArguments and adds any provided relocation mapping
// to the operation's files.
func (r *Runtime) AddRelocation(args ActionArguments) cnabaction.OperationConfigFunc {
	return func(op *driver.Operation) error {
		if len(args.BundleReference.RelocationMap) > 0 {
			b, err := json.MarshalIndent(args.BundleReference.RelocationMap, "", "    ")
			if err != nil {
				return fmt.Errorf("error marshaling relocation mapping file: %w", err)
			}

			op.Files["/cnab/app/relocation-mapping.json"] = string(b)

			// If the invocation image is present in the relocation mapping, we need
			// to update the operation and set the new image reference. Unfortunately,
			// the relocation mapping is just reference => reference, so there isn't a
			// great way to check for the invocation image.
			if mappedInvo, ok := args.BundleReference.RelocationMap[op.Image.Image]; ok {
				op.Image.Image = mappedInvo
			}
		}
		return nil
	}
}

func (r *Runtime) Execute(ctx context.Context, args ActionArguments) error {
	// Check if we've been asked to stop before executing long blocking calls
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		ctx, log := tracing.StartSpan(ctx,
			attribute.String("action", args.Action),
			attribute.Bool("allowDockerHostAccess", args.AllowDockerHostAccess),
			attribute.String("driver", args.Driver))
		defer log.EndSpan()
		args.BundleReference.AddToTrace(ctx)
		args.Installation.AddToTrace(ctx)

		if args.Action == "" {
			return log.Error(errors.New("action is required"))
		}

		b, err := r.ProcessBundle(ctx, args.BundleReference.Definition)
		if err != nil {
			return log.Error(err)
		}

		currentRun, err := r.CreateRun(ctx, args, b)
		if err != nil {
			return log.Error(err)
		}

		// Validate the action
		if _, err := b.GetAction(currentRun.Action); err != nil {
			return log.Error(fmt.Errorf("invalid action '%s' specified for bundle %s: %w", currentRun.Action, b.Name, err))
		}

		creds, err := r.loadCredentials(ctx, b, args)
		if err != nil {
			return log.Error(fmt.Errorf("not load credentials: %w", err))
		}

		log.Debugf("Using runtime driver %s\n", args.Driver)
		driver, err := r.newDriver(args.Driver, args)
		if err != nil {
			return log.Error(fmt.Errorf("unable to instantiate driver: %w", err))
		}

		a := cnabaction.New(driver)
		a.SaveLogs = args.PersistLogs

		if currentRun.ShouldRecord() {
			err = r.SaveRun(ctx, args.Installation, currentRun, cnab.StatusRunning)
			if err != nil {
				return log.Error(fmt.Errorf("could not save the pending action's status, the bundle was not executed: %w", err))
			}
		}

		opResult, result, err := a.Run(currentRun.ToCNAB(), creds.ToCNAB(), r.ApplyConfig(ctx, args)...)

		if currentRun.ShouldRecord() {
			if err != nil {
				err = r.appendFailedResult(ctx, err, currentRun)
				return log.Error(fmt.Errorf("failed to record that %s for installation %s failed: %w", args.Action, args.Installation.Name, err))
			}
			return r.SaveOperationResult(ctx, opResult, args.Installation, currentRun, currentRun.NewResultFrom(result))
		}

		if err != nil {
			return log.Error(fmt.Errorf("execution of %s for installation %s failed: %w", args.Action, args.Installation.Name, err))
		}

		return nil
	}
}

func (r *Runtime) CreateRun(ctx context.Context, args ActionArguments, b cnab.ExtendedBundle) (storage.Run, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// Create a record for the run we are about to execute
	var currentRun = args.Installation.NewRun(args.Action)
	currentRun.Bundle = b.Bundle
	currentRun.BundleReference = args.BundleReference.Reference.String()
	currentRun.BundleDigest = args.BundleReference.Digest.String()

	var err error
	extb := cnab.NewBundle(b.Bundle)
	currentRun.Parameters.Parameters, err = r.sanitizer.CleanRawParameters(ctx, args.Params, extb, currentRun.ID)
	if err != nil {
		return storage.Run{}, span.Error(err)
	}

	// TODO: Do not save secrets when the run isn't recorded
	currentRun.ParameterOverrides = storage.LinkSensitiveParametersToSecrets(currentRun.ParameterOverrides, extb, currentRun.ID)
	currentRun.CredentialSets = args.Installation.CredentialSets
	sort.Strings(currentRun.CredentialSets)

	currentRun.ParameterSets = args.Installation.ParameterSets
	sort.Strings(currentRun.ParameterSets)
	return currentRun, nil
}

// SaveRun with the specified status.
func (r *Runtime) SaveRun(ctx context.Context, installation storage.Installation, run storage.Run, status string) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debugf("saving action %s for %s installation with status %s", run.Action, installation, status)

	// update installation record to use run id ecoded parameters instead of
	// installation id
	installation.Parameters.Parameters = run.ParameterOverrides.Parameters
	err := r.installations.UpsertInstallation(ctx, installation)
	if err != nil {
		return span.Error(fmt.Errorf("error saving the installation record before executing the bundle: %w", err))
	}

	result := run.NewResult(status)
	err = r.installations.InsertRun(ctx, run)
	if err != nil {
		return span.Error(fmt.Errorf("error saving the installation run record before executing the bundle: %w", err))
	}

	err = r.installations.InsertResult(ctx, result)
	if err != nil {
		return span.Error(fmt.Errorf("error saving the installation status record before executing the bundle: %w", err))
	}

	return nil
}

// SaveOperationResult saves the ClaimResult and Outputs. The caller is
// responsible for having already persisted the claim itself, for example using
// SaveRun.
func (r *Runtime) SaveOperationResult(ctx context.Context, opResult driver.OperationResult, installation storage.Installation, run storage.Run, result storage.Result) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// TODO(carolynvs): optimistic locking on updates

	// Keep accumulating errors from any error returned from the operation
	// We must save the claim even when the op failed, but we want to report
	// ALL errors back.
	var bigerr *multierror.Error
	bigerr = multierror.Append(bigerr, opResult.Error)

	err := r.installations.InsertResult(ctx, result)
	if err != nil {
		bigerr = multierror.Append(bigerr, fmt.Errorf("error adding %s result for %s run of installation %s\n%#v: %w", result.Status, run.Action, installation, result, err))
	}

	installation.ApplyResult(run, result)
	err = r.installations.UpdateInstallation(ctx, installation)
	if err != nil {
		bigerr = multierror.Append(bigerr, fmt.Errorf("error updating installation record for %s\n%#v: %w", installation, installation, err))
	}

	for outputName, outputValue := range opResult.Outputs {
		output := result.NewOutput(outputName, []byte(outputValue))
		output, err = r.sanitizer.CleanOutput(ctx, output, cnab.ExtendedBundle{Bundle: run.Bundle})
		if err != nil {
			bigerr = multierror.Append(bigerr, fmt.Errorf("error sanitizing sensitive %s output for %s run of installation %s\n%#v: %w", output.Name, run.Action, installation, output, err))
		}
		err = r.installations.InsertOutput(ctx, output)
		if err != nil {
			bigerr = multierror.Append(bigerr, fmt.Errorf("error adding %s output for %s run of installation %s\n%#v: %w", output.Name, run.Action, installation, output, err))
		}
	}

	return bigerr.ErrorOrNil()
}

// appendFailedResult creates a failed result from the operation error and accumulates
// the error(s).
func (r *Runtime) appendFailedResult(ctx context.Context, opErr error, run storage.Run) error {
	saveResult := func() error {
		result := run.NewResult(cnab.StatusFailed)
		return r.installations.InsertResult(ctx, result)
	}

	resultErr := saveResult()

	// Accumulate any errors from the operation with the persistence errors
	return multierror.Append(opErr, resultErr).ErrorOrNil()
}
