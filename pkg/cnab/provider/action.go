package cnabprovider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/action"
	cnabaction "github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/driver"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Shared arguments for all CNAB actions
type ActionArguments struct {
	// Action to execute, e.g. install, upgrade.
	Action string

	// Name of the installation.
	Installation claims.Installation

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
}

func (r *Runtime) ApplyConfig(args ActionArguments) action.OperationConfigs {
	return action.OperationConfigs{
		r.SetOutput(),
		r.AddFiles(args),
		r.AddRelocation(args),
	}
}

func (r *Runtime) SetOutput() action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		op.Out = r.Out
		op.Err = r.Err
		return nil
	}
}

func (r *Runtime) AddFiles(args ActionArguments) action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		for k, v := range args.Files {
			op.Files[k] = v
		}

		// Add claim.json to file list as well, if exists
		claim, err := r.claims.GetLastRun(args.Installation.Namespace, args.Installation.Name)
		if err == nil {
			claimBytes, err := json.Marshal(claim)
			if err != nil {
				return errors.Wrapf(err, "could not marshal claim %s for installation %s", claim.ID, args.Installation)
			}
			op.Files[config.ClaimFilepath] = string(claimBytes)
		}

		return nil
	}
}

// AddRelocation operates on an ActionArguments and adds any provided relocation mapping
// to the operation's files.
func (r *Runtime) AddRelocation(args ActionArguments) action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		if len(args.BundleReference.RelocationMap) > 0 {
			b, err := json.MarshalIndent(args.BundleReference.RelocationMap, "", "    ")
			if err != nil {
				return errors.Wrapf(err, "error marshaling relocation mapping file")
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

func (r *Runtime) Execute(args ActionArguments) error {
	if args.Action == "" {
		return errors.New("action is required")
	}

	b, err := r.ProcessBundle(args.BundleReference.Definition)
	if err != nil {
		return err
	}

	if r.Debug {
		b.WriteTo(r.Err)
		fmt.Fprintln(r.Err)
	}

	// Create a record for the run we are about to execute
	var currentRun = args.Installation.NewRun(args.Action)
	currentRun.Bundle = b.Bundle
	currentRun.BundleReference = args.BundleReference.Reference.String()
	currentRun.BundleDigest = args.BundleReference.Digest.String()
	currentRun.Parameters = args.Params
	currentRun.CredentialSets = args.Installation.CredentialSets
	sort.Strings(currentRun.CredentialSets)
	currentRun.ParameterSets = args.Installation.ParameterSets
	sort.Strings(currentRun.ParameterSets)

	// Validate the action
	if _, err := b.GetAction(currentRun.Action); err != nil {
		return errors.Wrapf(err, "invalid action '%s' specified for bundle %s", currentRun.Action, b.Name)
	}

	creds, err := r.loadCredentials(b, args)
	if err != nil {
		return errors.Wrap(err, "could not load credentials")
	}

	fmt.Fprintf(r.Err, "Using runtime driver %s\n", args.Driver)
	driver, err := r.newDriver(args.Driver, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}

	a := cnabaction.New(driver)
	a.SaveLogs = true

	if currentRun.ShouldRecord() {
		err = r.SaveRun(args.Installation, currentRun, cnab.StatusRunning)
		if err != nil {
			return errors.Wrap(err, "could not save the pending action's status, the bundle was not executed")
		}
	}

	r.printDebugInfo(b, creds, args.Params)

	opResult, result, err := a.Run(currentRun.ToCNAB(), creds.ToCNAB(), r.ApplyConfig(args)...)

	if currentRun.ShouldRecord() {
		if err != nil {
			err = r.appendFailedResult(err, currentRun)
			return errors.Wrapf(err, "failed to record that %s for installation %s failed", args.Action, args.Installation.Name)
		}
		return r.SaveOperationResult(opResult, args.Installation, currentRun, currentRun.NewResultFrom(result))
	} else {
		return errors.Wrapf(err, "execution of %s for installation %s failed", args.Action, args.Installation.Name)
	}
}

// SaveRun with the specified status.
func (r *Runtime) SaveRun(installation claims.Installation, run claims.Run, status string) error {
	if r.Debug {
		fmt.Fprintf(r.Err, "saving action %s for %s installation with status %s\n", run.Action, installation, status)
	}
	err := r.claims.UpsertInstallation(installation)
	if err != nil {
		return errors.Wrap(err, "error saving the installation record before executing the bundle")
	}

	result := run.NewResult(status)
	err = r.claims.InsertRun(run)
	if err != nil {
		return errors.Wrap(err, "error saving the installation run record before executing the bundle")
	}

	err = r.claims.InsertResult(result)
	return errors.Wrap(err, "error saving the installation status record before executing the bundle")
}

// SaveOperationResult saves the ClaimResult and Outputs. The caller is
// responsible for having already persisted the claim itself, for example using
// SaveRun.
func (r *Runtime) SaveOperationResult(opResult driver.OperationResult, installation claims.Installation, run claims.Run, result claims.Result) error {
	// TODO(carolynvs): optimistic locking on updates

	// Keep accumulating errors from any error returned from the operation
	// We must save the claim even when the op failed, but we want to report
	// ALL errors back.
	var bigerr *multierror.Error
	bigerr = multierror.Append(bigerr, opResult.Error)

	err := r.claims.InsertResult(result)
	if err != nil {
		bigerr = multierror.Append(bigerr, errors.Wrapf(err, "error adding %s result for %s run of installation %s\n%#v", result.Status, run.Action, installation, result))
	}

	installation.ApplyResult(run, result)
	err = r.claims.UpdateInstallation(installation)
	if err != nil {
		bigerr = multierror.Append(bigerr, errors.Wrapf(err, "error updating installation record for %s\n%#v", installation, installation))
	}

	for outputName, outputValue := range opResult.Outputs {
		output := result.NewOutput(outputName, []byte(outputValue))
		err = r.claims.InsertOutput(output)
		if err != nil {
			bigerr = multierror.Append(bigerr, errors.Wrapf(err, "error adding %s output for %s run of installation %s\n%#v", output.Name, run.Action, installation, output))
		}
	}

	return bigerr.ErrorOrNil()
}

// appendFailedResult creates a failed result from the operation error and accumulates
// the error(s).
func (r *Runtime) appendFailedResult(opErr error, run claims.Run) error {
	saveResult := func() error {
		result := run.NewResult(cnab.StatusFailed)
		return r.claims.InsertResult(result)
	}

	resultErr := saveResult()

	// Accumulate any errors from the operation with the persistence errors
	return multierror.Append(opErr, resultErr).ErrorOrNil()
}

func (r *Runtime) printDebugInfo(b cnab.ExtendedBundle, creds secrets.Set, params map[string]interface{}) {
	if r.Debug {
		dump := &bytes.Buffer{}
		secrets := make([]string, 0, len(params)+len(creds))

		fmt.Fprintf(dump, "params:\n")
		for k, v := range params {
			if b.IsSensitiveParameter(k) {
				// TODO(carolynvs): When we consolidate our conversion logic of parameters into strings, let's use it here.
				// https://github.com/cnabio/cnab-go/issues/270
				secrets = append(secrets, fmt.Sprintf("%v", v))
			}
			fmt.Fprintf(dump, "  - %s: %v\n", k, v)
		}

		fmt.Fprintf(dump, "creds:\n")
		for k, v := range creds {
			secrets = append(secrets, fmt.Sprintf("%v", v))
			fmt.Fprintf(dump, "  - %s: %v\n", k, v)
		}

		r.Context.SetSensitiveValues(secrets)
		fmt.Fprintln(r.Err, dump.String())
	}
}
