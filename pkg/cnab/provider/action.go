package cnabprovider

import (
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/action"
	cnabaction "github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Shared arguments for all CNAB actions
type ActionArguments struct {
	// Action to execute, e.g. install, upgrade.
	Action string

	// Namespace of the installation.
	Namespace string

	// Name of the installation.
	Installation string

	// Either a filepath to the bundle or the name of the bundle.
	BundlePath string

	// BundleReference is the OCI reference of the bundle.
	BundleReference string

	// BundleDigest is the digest of the pulled bundle.
	BundleDigest string

	// Additional files to copy into the bundle
	// Target Path => File Contents
	Files map[string]string

	// Params is the set of user-specified parameter values to pass to the bundle.
	Params map[string]string

	// Either a filepath to a credential file or the name of a set of a credentials.
	CredentialIdentifiers []string

	// Driver is the CNAB-compliant driver used to run bundle actions.
	Driver string

	// Path to an optional relocation mapping file
	RelocationMapping string

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
		claim, err := r.claims.GetLastRun(args.Namespace, args.Installation)
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
		if args.RelocationMapping != "" {
			b, err := r.FileSystem.ReadFile(args.RelocationMapping)
			if err != nil {
				return errors.Wrap(err, "unable to add relocation mapping")
			}
			op.Files["/cnab/app/relocation-mapping.json"] = string(b)
			var reloMap relocation.ImageRelocationMap
			err = json.Unmarshal(b, &reloMap)
			// If the invocation image is present in the relocation mapping, we need
			// to update the operation and set the new image reference. Unfortunately,
			// the relocation mapping is just reference => reference, so there isn't a
			// great way to check for the invocation image.
			if mappedInvo, ok := reloMap[op.Image.Image]; ok {
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

	var b bundle.Bundle
	var err error

	if args.BundlePath != "" {
		b, err = r.ProcessBundleFromFile(args.BundlePath)
		if err != nil {
			return err
		}
	}

	// Load the installation
	installation, err := r.claims.GetInstallation(args.Namespace, args.Installation)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound{}) {
			// This may not exist because it's an invoked command running before install
			installation = claims.NewInstallation(args.Namespace, args.Installation)
		} else {
			return errors.Wrapf(err, "could not retrieve the installation record")
		}
	}

	lastRun, err := r.claims.GetLastRun(args.Namespace, args.Installation)
	if err != nil {
		// Only install and stateless actions can execute without an initial installation
		if !(args.Action == cnab.ActionInstall || b.Actions[args.Action].Stateless) {
			instNotFound := storage.ErrNotFound{Collection: claims.CollectionInstallations}
			return errors.Wrapf(instNotFound, "Installation not found: %s. Only stateless actions, such as dry-run, can execute against a bundle that hasn't been installed yet", args.Installation)
		}
	}

	// If the user didn't override the bundle definition, use the one
	// from the existing claim
	if lastRun.ID != "" && args.BundlePath == "" {
		b, err = r.ProcessBundle(lastRun.Bundle)
		if err != nil {
			return err
		}
	}
	if r.Debug {
		b.WriteTo(r.Err)
		fmt.Fprintln(r.Err)
	}

	params, err := r.loadParameters(b, args)
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	// Create a record for the run we are about to execute
	var currentRun = installation.NewRun(args.Action)
	currentRun.Bundle = b
	currentRun.BundleReference = args.BundleReference
	currentRun.BundleDigest = args.BundleDigest
	currentRun.Parameters = params

	// Validate the action
	if _, err := currentRun.Bundle.GetAction(currentRun.Action); err != nil {
		return errors.Wrapf(err, "invalid action '%s' specified for bundle %s", currentRun.Action, currentRun.Bundle.Name)
	}

	creds, err := r.loadCredentials(currentRun.Bundle, args)
	if err != nil {
		return errors.Wrap(err, "could not load credentials")
	}

	driver, err := r.newDriver(args.Driver, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}

	a := cnabaction.New(driver)
	a.SaveLogs = true

	if currentRun.ShouldRecord() {
		err = r.SaveRun(installation, currentRun, cnab.StatusRunning)
		if err != nil {
			return errors.Wrap(err, "could not save the pending action's status, the bundle was not executed")
		}
	}

	r.printDebugInfo(creds, params)

	opResult, result, err := a.Run(currentRun.ToCNAB(), creds.ToCNAB(), r.ApplyConfig(args)...)

	if currentRun.ShouldRecord() {
		if err != nil {
			err = r.appendFailedResult(err, currentRun)
			return errors.Wrapf(err, "failed to %s the bundle", args.Action)
		}
		return r.SaveOperationResult(opResult, installation, currentRun, currentRun.NewResultFrom(result))
	} else {
		return errors.Wrapf(err, "failed to %s the bundle", args.Action)
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

func (r *Runtime) printDebugInfo(creds secrets.Set, params map[string]interface{}) {
	if r.Debug {
		// only print out the names of the credentials, not the contents, cuz they big and sekret
		credKeys := make([]string, 0, len(creds))
		for k := range creds {
			credKeys = append(credKeys, k)
		}
		// param values may also be sensitive, so just print names
		paramKeys := make([]string, 0, len(params))
		for k := range params {
			paramKeys = append(paramKeys, k)
		}
		fmt.Fprintf(r.Err, "params: %v\ncreds: %v\n", paramKeys, credKeys)
	}
}
