package cnabprovider

import (
	"get.porter.sh/porter/pkg/manifest"
	"github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
)

func (r *Runtime) Upgrade(args ActionArguments) error {
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
		return errors.Wrapf(err, "could not load installation %s", args.Installation)
	}

	// If the user didn't override the bundle definition, use the one
	// from the existing claim
	if args.BundlePath == "" {
		b = existingClaim.Bundle
	}

	params, err := r.loadParameters(b, args.Params, args.ParameterSets, string(manifest.ActionUpgrade))
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	c, err := existingClaim.NewClaim(claim.ActionUpgrade, b, params)
	if err != nil {
		return err
	}

	creds, err := r.loadCredentials(c.Bundle, args.CredentialIdentifiers)
	if err != nil {
		return errors.Wrap(err, "could not load credentials")
	}

	driver, err := r.newDriver(args.Driver, args.Installation, args)
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
		return errors.Wrap(err, "failed to upgrade the bundle")
	}

	return a.SaveOperationResult(opResult, c, result)
}
