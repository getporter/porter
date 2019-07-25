package cnabprovider

import (
	"fmt"

	"github.com/deislabs/cnab-go/action"
	"github.com/pkg/errors"
)

func (d *Duffle) Upgrade(args ActionArguments) error {
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
		claim.Parameters, err = d.loadParameters(&claim, args.Params)
		if err != nil {
			return errors.Wrap(err, "invalid parameters")
		}
	}

	driver, err := d.newDriver(args.Driver, claim.Name)
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

	// Upgrade and capture error
	runErr := i.Run(&claim, creds, d.Out)

	// Add/update the outputs section of a claim and capture error
	err = d.WriteClaimOutputs(&claim)

	// ALWAYS write out a claim, even if the upgrade fails
	saveErr := claims.Store(claim)
	if runErr != nil {
		return errors.Wrap(runErr, "failed to upgrade the bundle")
	}
	if err != nil {
		return errors.Wrap(err, "failed to write outputs to the claim")
	}
	return errors.Wrap(saveErr, "failed to record the upgrade for the bundle")
}
