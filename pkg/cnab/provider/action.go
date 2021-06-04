package cnabprovider

import (
	"encoding/json"
	"fmt"

	cnabaction "github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/bundle"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Shared arguments for all CNAB actions
type ActionArguments struct {
	// Action to execute, e.g. install, upgrade.
	Action string

	// Name of the installation.
	Installation string

	// Either a filepath to the bundle or the name of the bundle.
	BundlePath string

	// BundleReference is the OCI reference of the bundle.
	BundleReference string

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
		claim, err := r.claims.ReadLastClaim(args.Installation)
		if err == nil {
			claimBytes, err := yaml.Marshal(claim)
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
		b, err = r.ProcessBundle(args.BundlePath)
		if err != nil {
			return err
		}
	}

	existingClaim, err := r.claims.ReadLastClaim(args.Installation)
	if err != nil {
		// Only install and stateless actions can execute without an initial installation
		if !(args.Action == claim.ActionInstall || b.Actions[args.Action].Stateless) {
			return errors.Wrapf(err, "could not load installation %s", args.Installation)
		}
	}

	// If the user didn't override the bundle definition, use the one
	// from the existing claim
	if existingClaim.ID != "" && args.BundlePath == "" {
		b = existingClaim.Bundle
	}

	params, err := r.loadParameters(b, args)
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	var c claim.Claim
	if existingClaim.ID == "" {
		c, err = claim.New(args.Installation, args.Action, b, params)
	} else {
		c, err = existingClaim.NewClaim(args.Action, b, params)
	}
	if err != nil {
		return err
	}

	c.BundleReference = args.BundleReference

	// Validate the action we are about to perform
	err = c.Validate()
	if err != nil {
		return err
	}

	creds, err := r.loadCredentials(c.Bundle, args)
	if err != nil {
		return errors.Wrap(err, "could not load credentials")
	}

	driver, err := r.newDriver(args.Driver, args.Installation, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}

	a := cnabaction.New(driver, r.claims)
	a.SaveAllOutputs = true
	a.SaveLogs = true

	modifies, err := c.IsModifyingAction()
	if err != nil {
		return err
	}

	// Only record runs that modify the bundle, e.g. don't save "logs" or "dry-run"
	// In theory a custom action shouldn't ever have modifies AND stateless
	// (which creates a temp claim) but just in case, if it does modify, we must
	// persist.
	shouldPersistClaim := func() bool {
		stateless := false
		if customAction, ok := c.Bundle.Actions[args.Action]; ok {
			stateless = customAction.Stateless
		}
		return modifies && !stateless
	}()

	if shouldPersistClaim {
		err = a.SaveInitialClaim(c, claim.StatusRunning)
		if err != nil {
			return err
		}
	}

	r.printDebugInfo(creds, params)

	opResult, result, err := a.Run(c, creds, r.ApplyConfig(args)...)

	if shouldPersistClaim {
		if err != nil {
			err = r.appendFailedResult(err, c)
			return errors.Wrapf(err, "failed to %s the bundle", args.Action)
		}
		return a.SaveOperationResult(opResult, c, result)
	} else {
		return errors.Wrapf(err, "failed to %s the bundle", args.Action)
	}
}

// appendFailedResult creates a failed result from the operation error and accumulates
// the error(s).
func (r *Runtime) appendFailedResult(opErr error, c claim.Claim) error {
	saveResult := func() error {
		result, err := c.NewResult(claim.StatusFailed)
		if err != nil {
			return err
		}
		return r.claims.SaveResult(result)
	}

	resultErr := saveResult()

	// Accumulate any errors from the operation with the persistence errors
	return multierror.Append(opErr, resultErr).ErrorOrNil()
}

func (r *Runtime) printDebugInfo(creds valuesource.Set, params map[string]interface{}) {
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
