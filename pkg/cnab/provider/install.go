package cnabprovider

import (
	"github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/claim"
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
