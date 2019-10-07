package cnabprovider

import (
	"fmt"

	"github.com/deislabs/cnab-go/action"
	"github.com/deislabs/porter/pkg/manifest"
	"github.com/pkg/errors"
)

func (d *Runtime) Upgrade(args ActionArguments) error {
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
		c.Parameters, err = d.loadParameters(&c, args.Params, string(manifest.ActionUpgrade))
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

	// ALWAYS write out a claim, even if the upgrade fails
	saveErr := claims.Store(c)
	if runErr != nil {
		return errors.Wrap(runErr, "failed to upgrade the bundle")
	}
	return errors.Wrap(saveErr, "failed to record the upgrade for the bundle")
}
