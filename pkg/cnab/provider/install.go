package cnabprovider

import (
	"fmt"

	"github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func (r *Runtime) Install(args ActionArguments) error {
	b, err := r.ProcessBundle(args.BundlePath)
	if err != nil {
		return err
	}

	params, err := r.loadParameters(b, args.Params, args.ParameterSets, claim.ActionInstall)
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	c, err := claim.New(args.Installation, claim.ActionInstall, b, params)
	if err != nil {
		return errors.Wrap(err, "invalid bundle instance name")
	}

	creds, err := r.loadCredentials(b, args.CredentialIdentifiers)
	if err != nil {
		return errors.Wrap(err, "could not load credentials")
	}

	driver, err := r.newDriver(args.Driver, c.Installation, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}

	a := action.New(driver, r.claims)
	a.SaveAllOutputs = true

	err = a.SaveInitialClaim(c, claim.StatusRunning)
	if err != nil {
		return err
	}

	r.printDebugInfo(creds, params)

	opResult, result, err := a.Run(c, creds, r.ApplyConfig(args)...)
	if err != nil {
		err = r.appendFailedResult(err, c)
		return errors.Wrap(err, "failed to install the bundle")
	}

	return a.SaveOperationResult(opResult, c, result)
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
