package cnabprovider

import (
	"fmt"

	cnabaction "github.com/deislabs/cnab-go/action"
	"github.com/pkg/errors"
)

func (d *Duffle) Invoke(action string, args ActionArguments) error {
	claims := d.NewClaimStore()
	claim, err := claims.Read(args.Claim)
	if err != nil {
		return errors.Wrapf(err, "could not load claim %s", args.Claim)
	}

	if args.BundleIdentifier != "" {
		// TODO: handle resolving based on bundle name
		claim.Bundle, err = d.LoadBundle(args.BundleIdentifier, args.Insecure)
		if err != nil {
			return err
		}
	}

	if len(args.Params) > 0 {
		claim.Parameters, err = d.loadParameters(&claim, args.Params, action)
		if err != nil {
			return errors.Wrap(err, "invalid parameters")
		}
	}

	driver, err := d.newDriver(args.Driver, claim.Name)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}

	i := cnabaction.RunCustom{
		Action: action,
		Driver: driver,
	}

	creds, err := d.loadCredentials(claim.Bundle, args.CredentialIdentifiers)
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
		paramKeys := make([]string, 0, len(claim.Parameters))
		for k := range claim.Parameters {
			paramKeys = append(paramKeys, k)
		}
		fmt.Fprintf(d.Err, "invoking bundle %s (%s) with action %s as %s\n\tparams: %v\n\tcreds: %v\n", claim.Bundle.Name, args.BundleIdentifier, action, claim.Name, paramKeys, credKeys)
	}

	// Run the action and ALWAYS write out a claim, even if the action fails
	err = i.Run(&claim, creds, d.Out)
	saveErr := claims.Store(claim)
	if err != nil {
		return errors.Wrap(err, "failed to invoke the bundle")
	}
	return errors.Wrap(saveErr, "failed to record the updated claim for the bundle")
}
