package cnabprovider

import (
	"fmt"

	"github.com/deislabs/cnab-go/action"
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/porter/pkg/config"
	"github.com/pkg/errors"
)

func (d *Duffle) Uninstall(args ActionArguments) error {
	// TODO: this entire function should be exposed in a duffle sdk package e.g. duffle.Install
	// we shouldn't be reimplementing calling all these functions all over again

	claims, err := d.NewClaimStore()
	if err != nil {
		return errors.Wrapf(err, "could not access claim store")
	}
	c, err := claims.Read(args.Claim)
	if err != nil {
		// Yay! It's already gone
		if err == claim.ErrClaimNotFound {
			return nil
		}
		return errors.Wrapf(err, "could not load claim %s", args.Claim)
	}

	if args.BundlePath != "" {
		// TODO: if they installed an insecure bundle, do they really need to do --insecure again to uninstall it?
		c.Bundle, err = d.LoadBundle(args.BundlePath, args.Insecure)
		if err != nil {
			return err
		}
	}

	if len(args.Params) > 0 {
		c.Parameters, err = d.loadParameters(&c, args.Params, string(config.ActionUninstall))
		if err != nil {
			return errors.Wrap(err, "invalid parameters")
		}
	}

	driver, err := d.newDriver(args.Driver, c.Name, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}
	i := action.Uninstall{
		Driver:          driver,
		OperationConfig: args.ApplyFiles(),
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
		fmt.Fprintf(d.Err, "uninstalling bundle %s (%s) as %s\n\tparams: %v\n\tcreds: %v\n", c.Bundle.Name, args.BundlePath, c.Name, paramKeys, credKeys)
	}

	err = i.Run(&c, creds, d.Out)
	if err != nil {
		return errors.Wrap(err, "failed to uninstall the bundle")
	}

	err = claims.Delete(args.Claim)
	if err != nil {
		return errors.Wrap(err, "failed to remove the record of the bundle")
	}

	return nil
}
