package cnabprovider

import (
	"fmt"

	"github.com/deislabs/duffle/pkg/action"
	"github.com/pkg/errors"
)

type UpgradeArguments struct {
	ActionArguments
}

func (d *Duffle) Upgrade(args UpgradeArguments) error {
	// TODO: this entire function should be exposed in a duffle sdk package e.g. duffle.Upgrade
	// we shouldn't be reimplementing calling all these functions all over again

	claims := d.NewClaimStore()
	claim, err := claims.Read(args.Claim)
	if err != nil {
		return errors.Wrapf(err, "could not load claim %s", args.Claim)
	}

	if args.BundleIdentifier != "" {
		// TODO: handle resolving based on bundle name
		// TODO: if they installed an insecure bundle, do they really need to do --insecure again to upgrade it?
		claim.Bundle, err = d.LoadBundle(args.BundleIdentifier, args.Insecure)
		if err != nil {
			return err
		}
	}

	if len(args.Params) > 0 {
		claim.Parameters, err = d.loadParameters(claim.Bundle, args.Params)
		if err != nil {
			return errors.Wrap(err, "invalid parameters")
		}
	}

	driver, err := d.newDriver(args.Driver)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}
	i := action.Upgrade{
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
		fmt.Fprintf(d.Err, "upgrading bundle %s (%s) as %s\n\tparams: %v\n\tcreds: %v\n", claim.Bundle.Name, args.BundleIdentifier, claim.Name, paramKeys, credKeys)
	}

	// Upgrade and ALWAYS write out a claim, even if the upgrade fails
	err = i.Run(&claim, creds, d.Out)
	saveErr := claims.Store(claim)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade the bundle")
	}
	return errors.Wrap(saveErr, "failed to record the upgrade for the bundle")
}
