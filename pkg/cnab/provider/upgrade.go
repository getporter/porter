package cnabprovider

import (
	"fmt"

	"github.com/deislabs/cnab-go/action"
	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/config"
)

func (d *Runtime) Upgrade(args ActionArguments) error {
	// TODO: this entire function should be exposed in a duffle sdk package e.g. duffle.Upgrade
	// we shouldn't be reimplementing calling all these functions all over again

	claims, err := d.NewClaimStore()
	if err != nil {
		return errors.Wrapf(err, "could not access claim store")
	}
	c, err := claims.Read(args.Claim)
	if err != nil {
		return errors.Wrapf(err, "could not load bundle instance %s", args.Claim)
	}

	if args.BundlePath != "" {
		// TODO: if they installed an insecure bundle, do they really need to do --insecure again to upgrade it?
		c.Bundle, err = d.LoadBundle(args.BundlePath, args.Insecure)
		if err != nil {
			return err
		}
	}

	if len(args.Params) > 0 {
		c.Parameters, err = d.loadParameters(&c, args.Params, string(config.ActionUpgrade))
		if err != nil {
			return errors.Wrap(err, "invalid parameters")
		}
	}

	driver, err := d.newDriver(args.Driver, c.Name, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}
	i := action.Upgrade{
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
		fmt.Fprintf(d.Err, "upgrading bundle %s (%s) as %s\n\tparams: %v\n\tcreds: %v\n", c.Bundle.Name, args.BundlePath, c.Name, paramKeys, credKeys)
	}

	// Upgrade and capture error
	runErr := i.Run(&c, creds, d.ApplyConfig(args)...)

	// Add/update the outputs section of a claim and capture error
	err = d.WriteClaimOutputs(&c, string(config.ActionUpgrade))

	// ALWAYS write out a claim, even if the upgrade fails
	saveErr := claims.Store(c)
	if runErr != nil {
		return errors.Wrap(runErr, "failed to upgrade the bundle")
	}
	if err != nil {
		return errors.Wrap(err, "failed to write outputs to the bundle instance")
	}
	return errors.Wrap(saveErr, "failed to record the upgrade for the bundle")
}
