package cnabprovider

import (
	"fmt"

	cnabaction "github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// getClaim reads an instance from the runtime's instance storage. If one is not found, the bundle
// is examined to see if the action is stateless. If the action is stateless, we create a new, temporary, claim
// Returns a pointer to the claim, a flag to indicate if the claim is temporary, and an error if present.
func (d *Runtime) getClaim(bun *bundle.Bundle, actionName, claimName string) (*claim.Claim, bool, error) {
	c, err := d.storage.Read(claimName)
	if err != nil {
		if bun != nil {
			if action, ok := bun.Actions[actionName]; ok {
				if action.Stateless {
					c = claim.Claim{
						Name:   claimName,
						Bundle: bun,
					}
					return &c, true, nil
				}
			}
		}
		return nil, false, errors.Wrapf(err, "could not load claim %s", claimName)
	}
	return &c, false, nil
}

// writeClaim will attempt to store the provided claim if, and only if, the claim is not
// a temporary claim
func (d *Runtime) writeClaim(tempClaim bool, c *claim.Claim) error {
	if !tempClaim {
		return d.storage.Store(*c)
	}
	return nil
}

func (d *Runtime) Invoke(action string, args ActionArguments) error {

	var bun *bundle.Bundle
	var err error
	if args.BundlePath != "" {
		bun, err = d.LoadBundle(args.BundlePath, args.Insecure)
		if err != nil {
			return err
		}
	}

	c, isTemp, err := d.getClaim(bun, action, args.Claim)

	// Here we need to check this again
	// If provided, we should set the bundle on the claim accordingly
	if args.BundlePath != "" {
		c.Bundle = bun
	}

	c.Parameters, err = d.loadParameters(c, args.Params, action)
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	driver, err := d.newDriver(args.Driver, c.Name, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}

	i := cnabaction.RunCustom{
		Action: action,
		Driver: driver,
	}

	creds, err := d.loadCredentials(c.Bundle, args.CredentialIdentifiers)
	if err != nil {
		return errors.Wrap(err, "could not load credentials")
	}

	if d.Debug {
		// only print out the names of the credentials, not the contents, cuz they big and sekret
		credKeys := make([]string, 0, len(creds))
		for k := range creds {
			credKeys = append(credKeys, k)
		}
		// param values may also be sensitive, so just print names
		paramKeys := make([]string, 0, len(c.Parameters))
		for k := range c.Parameters {
			paramKeys = append(paramKeys, k)
		}
		fmt.Fprintf(d.Err, "invoking bundle %s (%s) with action %s as %s\n\tparams: %v\n\tcreds: %v\n", c.Bundle.Name, args.BundlePath, action, c.Name, paramKeys, credKeys)
	}

	var result *multierror.Error
	// Run the action and ALWAYS write out a claim, even if the action fails
	err = i.Run(c, creds, d.ApplyConfig(args)...)
	if err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to invoke the bundle"))
	}

	if !isTemp {
		// ALWAYS write out a claim, even if the action fails
		// We don't persist temporary claims generated for stateless actions.
		err = d.storage.Store(*c)
		if err != nil {
			result = multierror.Append(result, errors.Wrap(err, "failed to record the updated claim for the bundle"))
		}
	}
	return result.ErrorOrNil()
}
