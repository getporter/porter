package cnabprovider

import (
	cnabaction "github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
)

// getClaimForInvoke reads a claim from the runtime's claim storage. If one is not found, the bundle
// is examined to see if the action is stateless. If the action is stateless, we create a new, temporary, claim
// Returns a pointer to the claim, a flag to indicate if the claim is temporary, and an error if present.
func (r *Runtime) getClaimForInvoke(bun bundle.Bundle, actionName, installation string) (claim.Claim, bool, error) {
	c, err := r.claims.ReadLastClaim(installation)
	if err != nil {
		if action, ok := bun.Actions[actionName]; ok {
			if action.Stateless {
				c = claim.Claim{
					Action:       actionName,
					Installation: installation,
					Bundle:       bun,
				}
				return c, true, nil
			}
		}
		return claim.Claim{}, false, errors.Wrapf(err, "could not load installation %s", installation)
	}
	return c, false, nil
}

func (r *Runtime) Invoke(action string, args ActionArguments) error {
	var b bundle.Bundle
	var err error

	if args.BundlePath != "" {
		b, err = r.ProcessBundle(args.BundlePath)
		if err != nil {
			return err
		}
	}

	existingClaim, isTempClaim, err := r.getClaimForInvoke(b, action, args.Installation)
	if err != nil {
		return err
	}

	// If the user didn't override the bundle definition, use the one
	// from the existing claim
	if args.BundlePath == "" {
		b = existingClaim.Bundle
	}

	params, err := r.loadParameters(b, args.Params, args.ParameterSets, claim.ActionUpgrade)
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	// TODO: (carolynvs) this should be consolidated with get claim for invoke, or teh logic in that function simplified
	c, err := existingClaim.NewClaim(action, b, params)
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

	a := cnabaction.New(driver, r.claims)
	a.SaveAllOutputs = true

	modifies, err := c.IsModifyingAction()
	if err != nil {
		return err
	}

	// Only record runs that modify the bundle, e.g. don't save "logs" or "dry-run"
	// In theory a custom action shouldn't ever have modifies AND stateless
	// (which creates a temp claim) but just in case, if it does modify, we must
	// persist.
	shouldPersistClaim := modifies || !isTempClaim

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
			return errors.Wrap(err, "failed to invoke the bundle")
		}
		return a.SaveOperationResult(opResult, c, result)
	} else {
		return errors.Wrap(err, "failed to invoke the bundle")
	}
}
